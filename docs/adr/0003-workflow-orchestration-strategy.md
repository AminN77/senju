# ADR-0003: Workflow orchestration strategy (custom worker queue first)

- **Status:** Accepted
- **Date:** 2026-04-10
- **Owners:** Platform team
- **Decision scope:** How pipeline stages are orchestrated from MVP to scale-up.

## Context

The platform executes multi-stage genomic workflows with dependencies:

1. QC (FastQC)
2. Alignment (BWA + SAMtools)
3. Variant calling (GATK)

MVP requires fast iteration and full visibility into stage behavior. Later phases may need richer scheduling and DAG-level controls.

## Decision

Use a **custom Go worker orchestration model** for MVP, backed by queue messages and explicit stage-state tracking in Postgres.

Design orchestration boundaries so the system can migrate to **Argo Workflows** or **Apache Airflow** when complexity or scale warrants.

## Alternatives considered

1. **Argo Workflows from day one**
   - Pros: strong DAG orchestration, cloud-native scaling.
   - Cons: high upfront complexity for MVP.
2. **Apache Airflow from day one**
   - Pros: mature scheduling and ecosystem.
   - Cons: heavier operational footprint and Python-centric orchestration stack.
3. **Pure shell-script orchestration**
   - Pros: fastest initial scripting.
   - Cons: poor reliability, observability, and maintainability for growth.

## Trade-offs

- **Gain:** rapid implementation and high control over execution semantics.
- **Cost:** custom orchestration logic must be maintained.
- **Constraint:** careful interface design is needed to avoid migration lock-in.

## Risks

- Orchestrator code could grow into brittle, implicit workflow logic.
- Retry/checkpoint semantics may become inconsistent across stages.

## Mitigations

- Enforce explicit stage contracts (inputs, outputs, exit semantics).
- Use a state machine model for valid transitions.
- Add checkpoint metadata and idempotent stage execution rules.
- Keep workflow definitions config-driven where possible.

## Migration path

1. Introduce `workflow_engine` interface and adapter boundary.
2. Externalize workflow definitions to declarative config.
3. Mirror one production workflow in Argo/Airflow as a pilot.
4. Cut over workflow-by-workflow with validation parity checks.

## Consequences

The project can ship MVP quickly with full control while preserving a credible path to enterprise orchestration platforms.
