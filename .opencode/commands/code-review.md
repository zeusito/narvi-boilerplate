---
description: "Perform a code review focusing on code quality"
---

Review the full diff and nearby code. Find real bugs, regressions, and unnecessary complexity. Reuse existing patterns, rank issues by severity, filter false positives, apply DRY/KISS refactors, fix, test, and re-review until clean.

## Reviewer Guidance

- The diff is your primary input — stay focused on what changed
- Read surrounding code in changed files to understand context
- Distinguish between new issues introduced by the diff and pre-existing issues
- Flag pre-existing issues if discovered
- Ask clarifying questions when the intent of a change is unclear
- Reference the specific file and change from the diff in each finding
- Recommendations target the principled long-term solution. Do not default to the minimal-diff resolution
- Classify findings by severity (critical, high, medium, low, info):
  - Not every review category will produce findings at every severity level
  - Use the levels that fit rather than forcing findings into categories that do not apply
- It is acceptable to find no issues. If the changes are well-implemented, say so. Do not manufacture findings to justify the review

### 1. Testing Review

Purpose: Evaluate test quality, coverage, and the testability of the subject.

- Coverage Depth: Assessing whether tests verify the logic under review across a representative range of scenarios.
- Suite Gating: Verifying that the suite runs on every proposed change and can block merge when it fails.
- Test Quality: Ensuring tests are readable, maintainable, and verify behaviour rather than implementation details.
- Testability: Verifying that the subject can be driven by automated tests through explicit seams and deterministic behaviour, rather than hidden state or live side effects.
- Test Isolation: Verifying that isolated tests do not share state or depend on ambient environment, and that tests which use real dependencies declare that choice rather than leaking it.
- Flakiness Prevention: Identifying tests that may fail intermittently due to timing or environmental factors.
- Contract Testing: Verifying that service interfaces match consumer expectations and do not break downstream integrations.
- Performance and Load Test Coverage: Assessing whether critical paths have tests under realistic load.
- End-to-End Testing: Assessing whether critical user journeys across service boundaries are covered by end-to-end tests.
- Chaos Test Coverage: Assessing whether controlled failure injection covers resilience under adverse conditions.
- Security Test Coverage: Assessing whether tests cover authentication, authorisation, input validation boundaries, and known attack patterns relevant to the reviewed subject.
- Non-Deterministic and Eval Behaviour: Assessing how tests handle non-deterministic model or agent output, including seeds, fixtures, eval harnesses, and tolerance bands that keep signal without masking regressions.

### 2. Security Review

Purpose: Identify vulnerabilities, security weaknesses, and potential attack vectors.

- Authentication and Authorisation: Verifying the integrity of identity verification and the strict enforcement of access boundaries across all layers.
- Session Management: Reviewing the lifecycle and security properties of user sessions and tokens to prevent hijacking or unauthorised reuse.
- Privilege Escalation: Analysing logic for flaws that could allow a user to perform actions beyond their intended permission level.
- Insecure Direct Object References: Verifying that access to resources by identifier enforces authorisation checks rather than relying on obscurity.
- Tenant Isolation: Verifying that data stores, caches, queues, background jobs, and search indices enforce tenant boundaries so no request can reach another tenant's records.
- Cryptography Usage: Evaluating the implementation of cryptographic primitives to ensure the use of proven, industry-standard protocols.
- Input Validation and Sanitisation: Ensuring all untrusted data is validated and cleaned to prevent injection and manipulation attacks.
- Excessive Data Exposure: Verifying that API and service responses return only fields the caller needs, and that debug, internal, or sensitive attributes are not leaked through over-fetch or verbose error payloads.
- CORS and CSRF Protection: Verifying that cross-origin policies and request forgery protections are correctly configured.
- Rate Limiting: Assessing the system's resilience against automated abuse, brute-force attempts, and resource exhaustion.
- Path Traversal: Ensuring that file and resource pathing logic cannot be manipulated to access restricted areas.
- Server-Side Request Forgery: Ensuring server-side requests cannot be manipulated to access internal resources or unintended external targets.
- Mass Assignment: Verifying that object binding from external input does not allow modification of unintended fields or properties.
- File Upload Security: Ensuring uploaded files are validated for content type, scanned for malicious content, and stored in isolated locations.
- Deserialisation Safety: Verifying that the conversion of data formats into objects does not introduce execution risks.
- Webhook and Callback Verification: Ensuring inbound callbacks from external systems are authenticated by signature and protected against replay.
- Time-of-Check to Time-of-Use: Identifying races where a permission, existence, or integrity check is separated from use so an attacker can change the resource in the gap.