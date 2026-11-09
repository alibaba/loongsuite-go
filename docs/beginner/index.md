# Beginner's Guide: Start Here

This guide is for readers who are new to Go instrumentation or OpenTelemetry. You don't need to know about compilers, ASTs or the OpenTelemetry SDK. Being able to read a simple Go function is enough.

The [User Guide](../index.md) shows how to use `otel`, while the [Developer Guide](../dev/overview.md) and [Hacking Guide](../hacking/overview.md) explain rules and internals. This guide covers the basics you need before reading them: each page answers one question with a small example, then points to the official page to read next.

## After this guide, you can answer

- If a program already returns the right result, why add "observation code" to it?
- When you prefix `go build` with `otel`, when and how does the extra code get added?
- What is a rule responsible for, and what is a hook responsible for?
- Why does the build need to download and prepare new dependencies after instrumentation?
- How does one request become a Trace made of multiple Spans?

## Reading order

| Step | Page | Question |
| --- | --- | --- |
| 1 | [Why Instrument](./why-instrument.md) | Why isn't a return value enough? |
| 2 | [Packages, Modules and go.mod](./modules.md) | How does Go find the code to compile? |
| 3 | [Build Time and Run Time](./build-and-run.md) | Where does `otel` step in? |
| 4 | [Rules and Hooks](./rules-and-hooks.md) | How does the tool know what to change and what to add? |
| 5 | [Why Add Dependencies](./dependencies.md) | Why does one extra call require new dependencies? |
| 6 | [Traces and Spans](./traces-and-spans.md) | How do separate records form one request? |
| 7 | [Hands-on: Your First Custom Hook](./first-hook.md) | How do you prove the hook really ran? |

## Before you start

- A Go version supported by your `otel` release; see [Compatibility](../user/compatibility.md) for the Go and OpenTelemetry requirements.
- Steps 1 to 6 are reading only; their examples run with a plain `go run`.
- Step 7 requires the `otel` tool. See [Installation](../index.md#installation).

## Diagrams

Pages 1–6 include seven looping diagrams; page 2 has separate diagrams for packages and for go.mod. Reduced-motion and print modes show the final stage; the full sequence is also explained in the text.

## Community resources

The community-maintained [LoongSuite Go Guide](https://qinghua362330.github.io/loongsuite-go-guide/?lang=en) (English and Simplified Chinese) provides interactive diagrams and exercises. It is maintained independently of the official documentation and is optional for this guide.
