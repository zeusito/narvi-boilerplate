---
name: code-review
description: Perform a thorough code review focusing on code quality, correctness, test coverage, and security. Use when reviewing code changes, git diffs, PRs, or verifying bug fixes and refactors.
license: MIT
metadata:
  version: "1.0.0"
---

# Code Review

Review the full diff and nearby code. Find real bugs, regressions, and unnecessary complexity. Reuse existing patterns, rank issues by severity, filter false positives, apply DRY/KISS refactors, fix, test, and re-review until clean.

## Workflow

1. **Identify the Target**: Inspect the change set (e.g. `git status`, `git diff`, branch diff against base, or user-specified files/commits).
2. **Read Surrounding Context**: Read beyond the diff lines in changed files to understand module architecture, invariants, and existing patterns.
3. **Evaluate Core Quality & Bugs**: Identify logic bugs, regressions, edge conditions, nullability, concurrency races, error handling paths, and state management flaws.
4. **Evaluate Testing & Verification**: Assess test quality, coverage of new edge cases, isolation, and testability.
5. **Evaluate Security**: Check against common attack vectors (authentication, injection, sanitization, isolation, data exposure).
6. **Classify Severity**: Categorize findings by severity (Critical, High, Medium, Low, Info) referencing exact files and line numbers.
7. **Produce Report**: Present findings inline with clear, principled recommendations for long-term resolution.

## Reviewer Guidance

- The diff is your primary input — stay focused on what changed.
- Read surrounding code in changed files to understand context.
- Distinguish between new issues introduced by the diff and pre-existing issues.
- Flag pre-existing issues if discovered (label clearly as pre-existing).
- Ask clarifying questions when the intent of a change is unclear.
- Reference the specific file and change from the diff in each finding (file path and line number).
- Recommendations target the principled long-term solution. Do not default to the minimal-diff resolution.
- Classify findings by severity (critical, high, medium, low, info):
  - Not every review category will produce findings at every severity level.
  - Use the levels that fit rather than forcing findings into categories that do not apply.
- It is acceptable to find no issues. If the changes are well-implemented, say so. Do not manufacture findings to justify the review.
- Write "None" for any severity tier where no findings exist.

## Review Scope

### 1. Code Correctness & Quality
- **Bug & Regression Risk**: Edge conditions, nullability, concurrency races, error handling paths, unhandled promise rejections.
- **Simplicity & Maintainability**: Adherence to DRY/KISS, readable abstractions, reuse of established codebase conventions.
- **Resource Management**: Memory leaks, connection pooling, file handle cleanup, stream exhaustion.

### 2. Testing Review
- **Coverage Depth**: Assessing whether tests verify the logic under review across a representative range of scenarios.
- **Suite Gating**: Verifying that the suite runs on every proposed change and can block merge when it fails.
- **Test Quality**: Ensuring tests are readable, maintainable, and verify behaviour rather than implementation details.
- **Testability**: Verifying that the subject can be driven by automated tests through explicit seams and deterministic behaviour, rather than hidden state or live side effects.
- **Test Isolation**: Verifying that isolated tests do not share state or depend on ambient environment, and that tests which use real dependencies declare that choice rather than leaking it.
- **Flakiness Prevention**: Identifying tests that may fail intermittently due to timing or environmental factors.
- **Contract Testing**: Verifying that service interfaces match consumer expectations and do not break downstream integrations.
- **Performance and Load Test Coverage**: Assessing whether critical paths have tests under realistic load.
- **End-to-End Testing**: Assessing whether critical user journeys across service boundaries are covered by end-to-end tests.
- **Chaos Test Coverage**: Assessing whether controlled failure injection covers resilience under adverse conditions.
- **Security Test Coverage**: Assessing whether tests cover authentication, authorisation, input validation boundaries, and known attack patterns relevant to the reviewed subject.
- **Non-Deterministic and Eval Behaviour**: Assessing how tests handle non-deterministic model or agent output, including seeds, fixtures, eval harnesses, and tolerance bands that keep signal without masking regressions.

### 3. Security Review
- **Authentication and Authorisation**: Verifying the integrity of identity verification and the strict enforcement of access boundaries across all layers.
- **Session Management**: Reviewing the lifecycle and security properties of user sessions and tokens to prevent hijacking or unauthorised reuse.
- **Privilege Escalation**: Analysing logic for flaws that could allow a user to perform actions beyond their intended permission level.
- **Insecure Direct Object References**: Verifying that access to resources by identifier enforces authorisation checks rather than relying on obscurity.
- **Tenant Isolation**: Verifying that data stores, caches, queues, background jobs, and search indices enforce tenant boundaries so no request can reach another tenant's records.
- **Cryptography Usage**: Evaluating the implementation of cryptographic primitives to ensure the use of proven, industry-standard protocols.
- **Input Validation and Sanitisation**: Ensuring all untrusted data is validated and cleaned to prevent injection and manipulation attacks.
- **Excessive Data Exposure**: Verifying that API and service responses return only fields the caller needs, and that debug, internal, or sensitive attributes are not leaked through over-fetch or verbose error payloads.
- **CORS and CSRF Protection**: Verifying that cross-origin policies and request forgery protections are correctly configured.
- **Rate Limiting**: Assessing the system's resilience against automated abuse, brute-force attempts, and resource exhaustion.
- **Path Traversal**: Ensuring that file and resource pathing logic cannot be manipulated to access restricted areas.
- **Server-Side Request Forgery**: Ensuring server-side requests cannot be manipulated to access internal resources or unintended external targets.
- **Mass Assignment**: Verifying that object binding from external input does not allow modification of unintended fields or properties.
- **File Upload Security**: Ensuring uploaded files are validated for content type, scanned for malicious content, and stored in isolated locations.
- **Deserialisation Safety**: Verifying that the conversion of data formats into objects does not introduce execution risks.
- **Webhook and Callback Verification**: Ensuring inbound callbacks from external systems are authenticated by signature and protected against replay.
- **Time-of-Check to Time-of-Use**: Identifying races where a permission, existence, or integrity check is separated from use so an attacker can change the resource in the gap.

## Report Format

```markdown
## Code Review Summary

- **Scope / Target**: [Branch/commits/files reviewed]
- **Findings Count**: [e.g. 0 critical, 1 high, 2 medium, 0 low]

### Critical Findings
- [Findings representing serious bugs, data loss risk, or security vulnerabilities, or "None"]

### High Findings
- [Findings that introduce regressions or major security/testing gaps, or "None"]

### Medium Findings
- [Maintainability issues, edge cases, test quality gaps, or "None"]

### Low / Info
- [Minor code style, typos, or pre-existing observations, or "None"]

### Assessment & Next Steps
[Overall assessment of the changes and recommendations before merging]
```
