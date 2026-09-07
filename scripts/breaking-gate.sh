#!/usr/bin/env bash
# Shared backward-compatibility gate for nevr-proto.
#
# WHY THIS FILE EXISTS AT ALL, because a gate whose justification is missing gets
# deleted by the next person who questions it:
#
#   A gate is derived from WHAT THE PUSH DOES and WHAT A MISTAKE COSTS -- never
#   from the shape of the thing it sits on. "It is a git repo, so it should have
#   push blocks" is not a reason. The question is always: WHO IS DOWNSTREAM, AND
#   CAN I TAKE IT BACK?
#
#   In ~/src/spritz, pushing IS the safety mechanism -- committed work can be
#   resumed on another machine after a power cut, and the worst case is
#   fix-and-push-again. Blocking a push there would strand you on old code. So
#   that repo must NEVER block.
#
#   Here, pushing is PUBLICATION to a registry other repositories fetch. Loss is
#   not the risk; PROPAGATION is. A consumer that has already fetched a broken
#   module cannot be un-consumed by a follow-up commit. So this repo BLOCKS.
#
#   Same tool, opposite rulings, because the cost of a mistake is opposite.
#
# WHY CI AND THE PRE-PUSH HOOK BOTH CALL THIS ONE FILE: two gates that agree
# today and are maintained separately will disagree eventually, and the first
# anyone hears of it is a break that one of them waved through. They cannot drift
# if they are the same file.
#
# Usage:
#   scripts/breaking-gate.sh --target registry            [--commits RANGE]
#   scripts/breaking-gate.sh --target .git#branch=main    [--commits RANGE]
#   --offline-ok    treat "could not check" as a warning instead of a failure
set -uo pipefail

TARGET=""
RANGE=""
OFFLINE_OK=0
while [ $# -gt 0 ]; do
  case "$1" in
    --target)     TARGET="$2"; shift 2 ;;
    --commits)    RANGE="$2";  shift 2 ;;
    --offline-ok) OFFLINE_OK=1; shift ;;
    *) echo "breaking-gate: unknown argument: $1" >&2; exit 2 ;;
  esac
done
[ -n "$TARGET" ] || { echo "breaking-gate: --target is required" >&2; exit 2; }

if [ "$TARGET" = "registry" ]; then
  # --against-registry is a first-class flag in buf 1.72.0 and is MUTUALLY
  # EXCLUSIVE with --against. It compares against the latest commit on the
  # default branch in the registry, which is the only target that means anything
  # for a push to main: `--against .git#branch=main` on a push to main compares
  # main WITH ITSELF and has exactly one possible outcome.
  set -- buf breaking --against-registry
else
  set -- buf breaking --against "$TARGET"
fi

out="$("$@" 2>&1)"; rc=$?

# THREE OUTCOMES, NOT TWO. "Checked, it is fine" and "could not check" are
# different facts and only one of them is good news. buf separates them by exit
# code, measured on buf 1.72.0:
#     0   clean
#   100   violations
#     1   could not check  ("the server hosted at that remote is unavailable",
#                           "not_found: resource with name ... was not found")
case "$rc" in
  0)
    echo "breaking-gate: OK -- no backward-incompatible changes vs ${TARGET}."
    exit 0 ;;
  1)
    echo "breaking-gate: COULD NOT CHECK compatibility against ${TARGET}." >&2
    echo "$out" | sed 's/^/  /' >&2
    if [ "$OFFLINE_OK" = "1" ]; then
      echo "breaking-gate: NOT BLOCKING -- this is a local push and you may be offline." >&2
      echo "breaking-gate: NOTHING WAS VERIFIED. CI will check it against the registry." >&2
      exit 0
    fi
    echo "breaking-gate: FAILING -- CI must never publish on an unverified check." >&2
    exit 1 ;;
  100)
    : ;;  # violations -- fall through to attribution
  *)
    echo "breaking-gate: unexpected buf exit ${rc}" >&2
    echo "$out" | sed 's/^/  /' >&2
    exit 1 ;;
esac

echo "breaking-gate: BACKWARD-INCOMPATIBLE CHANGES vs ${TARGET}:" >&2
echo "$out" | sed 's/^/  /' >&2
echo >&2

# THE DELIBERATE-BREAK HATCH.
#
# `Breaking-Approved: <reason>` as a trailer on THE COMMIT THAT BREAKS.
#
# Why a trailer and not an environment variable: an env var waves through the
# ENTIRE PUSH. One push here carried nine commits. A trailer excuses only the
# commit carrying it, and leaves the justification permanently in the log where
# the next person can read WHY rather than discovering that someone once set a
# variable. Blast radius, not ceremony.
#
# Why `git interpret-trailers --parse` and not grep: grep matches the string
# anywhere in the message, including prose. Measured -- a commit body containing
# the line "Breaking-Approved: mentioned only in prose" in a middle paragraph is
# matched by `grep '^Breaking-Approved:'` and correctly IGNORED by
# interpret-trailers, which reads only the trailer block. A gate that can be
# tripped by prose is a gate that will be.
if [ -z "$RANGE" ]; then
  echo "breaking-gate: no commit range supplied, so no commit can be credited" >&2
  echo "breaking-gate: BLOCKED." >&2
  exit 1
fi

commits="$(git rev-list --reverse "$RANGE" 2>/dev/null)"
if [ -z "$commits" ]; then
  echo "breaking-gate: commit range '${RANGE}' is empty; BLOCKED." >&2
  exit 1
fi

unapproved=0
for c in $commits; do
  # Did THIS commit introduce a break? Compare it against its own parent, not
  # against the baseline -- against the baseline every commit after the breaking
  # one inherits the violation and gets blamed for it.
  if git rev-parse -q --verify "${c}^" >/dev/null 2>&1; then
    if buf breaking ".git#ref=${c}" --against ".git#ref=${c}^" >/dev/null 2>&1; then
      continue   # this commit broke nothing
    fi
  else
    continue     # root commit, nothing to compare against
  fi
  subject="$(git log -1 --format=%s "$c")"
  approval="$(git log -1 --format=%B "$c" | git interpret-trailers --parse | sed -n 's/^Breaking-Approved:[[:space:]]*//p')"
  if [ -n "$approval" ]; then
    echo "breaking-gate: APPROVED  ${c:0:8} ${subject}" >&2
    echo "breaking-gate:           reason: ${approval}" >&2
  else
    echo "breaking-gate: UNAPPROVED ${c:0:8} ${subject}" >&2
    unapproved=$((unapproved + 1))
  fi
done

if [ "$unapproved" -eq 0 ]; then
  echo >&2
  echo "breaking-gate: every breaking commit carries Breaking-Approved. Allowing." >&2
  exit 0
fi

cat >&2 <<MSG

breaking-gate: BLOCKED -- ${unapproved} commit(s) break compatibility without approval.

  This is not a formality. Publishing to buf.build/echotools/nevr-api is not
  reversible by a follow-up commit: a consumer that has already fetched the
  module cannot be un-consumed.

  If the break is intended, add the trailer TO THE COMMIT THAT BREAKS:

      git commit --amend --trailer "Breaking-Approved: <why this is safe now>"

  For a commit further back, rebase to reach it. The trailer excuses only the
  commit it is on -- that is the point.
MSG
exit 1
