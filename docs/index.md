---
layout: home

hero:
  name: protoc-gen-nats-micro
  text: Type-Safe NATS Microservices from Protobuf
  tagline: Write .proto files. Run buf generate. Get production-ready NATS services with discovery, load balancing, and streaming — zero config.
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
    - theme: alt
      text: API Reference
      link: /api/reference
    - theme: alt
      text: GitHub
      link: https://github.com/Toyz/protoc-gen-nats-micro

features:
  - title: Zero Configuration
    details: Service metadata, subjects, timeouts — all defined in your .proto files. Just run buf generate.
  - title: Type-Safe Everything
    details: Compile-time safety for requests, responses, errors, and interceptors across Go, TypeScript, and Python.
  - title: Streaming RPC
    details: Server-streaming, client-streaming, and bidirectional streaming over NATS pub/sub with typed wrappers.
  - title: KV and Object Store
    details: Auto-persist RPC responses to NATS KV or Object Store with configurable key templates.
  - title: Interceptors and Headers
    details: Full middleware support — logging, auth, tracing. Bidirectional header propagation on requests and responses.
  - title: Multi-Language
    details: Generate Go, TypeScript, and Python from the same proto definition. Same wire protocol everywhere.
---
