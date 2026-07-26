# Latency Measurement & Performance Analysis Report

This report presents the setup and performance metrics comparing async vs sync submission processing on the local distributed coding judge platform (**CodeLab**).

## 1. Setup Check Summary

* **Active Services**: Verified that the REST API Gateway (port `8080`), PostgreSQL DB (port `5432`), NATS JetStream (port `4222`), and the Judge Worker daemon are running.
* **Submission API Endpoints**:
  * `POST /api/submissions`: Asynchronously enqueues a coding submission to NATS and inserts a database record. (Protected by JWT Auth and CSRF middleware).
  * `GET /api/submissions/:id`: Returns status, execution metadata, and verdict (e.g. `PENDING`, `AC`, `WA`, `RE`).
* **Sandbox Execution Location**: Code resides in [backend/internal/sandbox/docker/runner.go](file:///D:/Projects/Personal/something/backend/internal/sandbox/docker/runner.go) which implements the `Compile`, `Run`, and `Cleanup` pipeline.
* **Problem**: Utilized the seeded `echo-input` problem.
* **Program**: Python code `print("Hello World")` submitted to the engine.

---

## 2. Benchmark Design & Implementation

We created a standalone Go CLI utility under [backend/cmd/benchmark/main.go](file:///D:/Projects/Personal/something/backend/cmd/benchmark/main.go). This utility:
1. Configures the database and loads the environment credentials.
2. Programmatically registers/seeds a verified user `benchmark@example.com` and generates a valid JWT access token.
3. Performs the following three latency tests sequentially:
   * **Measurement 1 — Async End-to-End Latency**: POSTs a submission, starts a timer, and polls `/api/submissions/:id` every 100ms until the verdict is no longer `PENDING`. Runs 20 times, tracking min/max/average. To comply with the `SubmissionRatePerSec` rate limiter, the script sleeps 3.1 seconds between submissions.
   * **Measurement 2 — API Response Time**: Measures only the HTTP POST request/response roundtrip to the `/api/submissions` endpoint (20 runs, sleeps 3.1s between iterations).
   * **Measurement 3 — Sync Baseline Simulation**: Connects to the Docker engine via the Go SDK and calls the sandbox runner's `Compile`, `Run`, and `Cleanup` methods directly. Runs 20 times, measuring total execution roundtrip.

---

## 3. Performance Results

The benchmark was executed against the active services. Below is the summary of the measurements:

| Metric | Result (Average) | Description |
| :--- | :--- | :--- |
| **Async API Response Time** | **29 ms** | Time taken by `/api/submissions` to respond and enqueue to NATS. |
| **Sync Sandbox Execution Time** | **624 ms** | Direct blocking Docker compilation, running, and workspace cleanup. |
| **End-to-End Verdict Latency** | **775 ms** | Total async roundtrip time (submission + queue + execution + polling). |

### Latency Summary

```
====================================================
LATENCY MEASUREMENT RESULTS
====================================================
Async API response time (submit endpoint): 29ms avg
Sync sandbox execution time (direct Docker call): 624ms avg
Improvement: 95.35% reduction in API blocking time
End-to-end verdict latency (async): 775ms avg
====================================================
```

---

## 4. Key Takeaways

1. **API Blocking Reduction**: Using an asynchronous NATS JetStream message queue yields a **95.35% reduction** in API blocking time (reducing response time from **624ms** to **29ms**). This is critical for scaling high-throughput request rates without running out of server worker pools.
2. **Worker & Sandbox Overhead**: The direct Docker container startup, mounting, execution, and cleanup takes **624ms** on average per execution. The extra ~150ms in the asynchronous end-to-end latency (**775ms**) accounts for queue ingress/egress transit lag, state persistence updates, and client polling intervals (which poll every 100ms).
