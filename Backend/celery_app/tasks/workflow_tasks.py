from Backend.api.routes import status
from Backend.celery_app import celery_app
from typing import Dict, List, Optional
from datetime import datetime
from Backend.data_layer.database.models.workflow import WorkflowStatus
from celery import chain, group
import enum
from Backend.celery_app.utils import async_to_sync


class StepStatus(str, enum.Enum):
    PENDING = "pending"
    ACTIVE = "active"
    COMPLETED = "completed"
    SKIPPED = "skipped"
    FAILED = "failed"


@celery_app.task(
    name="tasks.workflow_tasks.execute_workflow_step",
    queue="workflow",
    priority=6,
    retry_backoff=True,
    max_retries=3
)
def execute_workflow_step(
    workflow_id: int,
    step_id: int,
    input_data: Dict,
    user_id: int,
    execution_id: Optional[int] = None
) -> Dict:
    """
    Execute a single workflow step asynchronously with proper execution tracking.
    """
    return async_to_sync(_execute_workflow_step)(
        workflow_id, step_id, input_data, user_id, execution_id
    )


async def _execute_workflow_step(
    workflow_id: int,
    step_id: int,
    input_data: Dict,
    user_id: int,
    execution_id: Optional[int] = None
) -> Dict:
    """Async implementation of workflow step execution."""
    from Backend.data_layer.repositories.workflow_repository import WorkflowRepository
    from Backend.data_layer.database.connection import get_db

    async for session in get_db():
        try:
            repo = WorkflowRepository(session)

            # Create step execution record if execution_id is provided
            if execution_id:
                step_execution = await repo.create_step_execution(
                    execution_id=execution_id,
                    step_id=step_id,
                    status=StepStatus.PENDING
                )

            # TODO: Implement actual workflow step execution logic
            result = {
                "status": StepStatus.COMPLETED,
                "workflow_id": workflow_id,
                "step_id": step_id,
                "execution_id": execution_id,
                "result": "Step executed successfully",
                "timestamp": datetime.utcnow().isoformat()
            }

            # Update step execution status - FIX: Use StepStatus.COMPLETED instead of SUCCESS
            if execution_id:
                await repo.update_step_execution(
                    execution_id=execution_id,
                    step_id=step_id,
                    status=StepStatus.COMPLETED,  # Changed from SUCCESS to COMPLETED
                    result=result
                )

            await session.commit()
            return result

        except Exception as e:
            await session.rollback()
            result = {
                "status": StepStatus.FAILED,
                "workflow_id": workflow_id,
                "step_id": step_id,
                "execution_id": execution_id,
                "error": str(e),
                "timestamp": datetime.utcnow().isoformat()
            }

            # Update step execution status on failure
            if execution_id:
                try:
                    await repo.update_step_execution(
                        execution_id=execution_id,
                        step_id=step_id,
                        status=StepStatus.FAILED,
                        error=str(e)
                    )
                    await session.commit()
                except Exception:
                    await session.rollback()

            return result


@celery_app.task(
    name="Backend.celery_app.tasks.workflow_tasks.create_workflow_task",
    queue="workflow",
    priority=5,
    retry_backoff=True,
    max_retries=3
)
def create_workflow_task(workflow_data: Dict) -> Dict:
    """Create a new workflow in the database."""
    return async_to_sync(_create_workflow_task)(workflow_data)


async def _create_workflow_task(workflow_data: Dict) -> Dict:
    """Async implementation of workflow creation."""
    from Backend.data_layer.repositories.workflow_repository import WorkflowRepository
    from Backend.data_layer.database.connection import get_db

    async for session in get_db():
        try:
            repo = WorkflowRepository(session)
            workflow = await repo.create_workflow(**workflow_data)
            await session.commit()
            return {"id": workflow.id, "status": workflow.status.value}
        except Exception as e:
            await session.rollback()
            raise  # Let the outer function handle retries


@celery_app.task(
    name="Backend.celery_app.tasks.workflow_tasks.delete_workflow_task",
    queue="workflow",
    priority=5,
    retry_backoff=True,
    max_retries=3
)
def delete_workflow_task(workflow_id: int) -> Dict:
    """Delete a workflow from the database."""
    return async_to_sync(_delete_workflow_task)(workflow_id)


async def _delete_workflow_task(workflow_id: int) -> Dict:
    """Async implementation of workflow deletion."""
    from Backend.data_layer.repositories.workflow_repository import WorkflowRepository
    from Backend.data_layer.database.connection import get_db

    async for session in get_db():
        try:
            repo = WorkflowRepository(session)
            await repo.delete_workflow(workflow_id)
            await session.commit()
            return {"status": "success", "deleted_id": workflow_id}
        except Exception as e:
            await session.rollback()
            raise  


@celery_app.task(
    name="Backend.celery_app.tasks.workflow_tasks.update_workflow_task",
    queue="workflow",
    priority=5,
    retry_backoff=True,
    max_retries=3
)
def update_workflow_task(workflow_id: int, updates: Dict) -> Dict:
    """Update an existing workflow in the database."""
    return async_to_sync(_update_workflow_task)(workflow_id, updates)


async def _update_workflow_task(workflow_id: int, updates: Dict) -> Dict:
    """Async implementation of workflow update."""
    from Backend.data_layer.repositories.workflow_repository import WorkflowRepository
    from Backend.data_layer.database.connection import get_db

    async for session in get_db():
        try:
            repo = WorkflowRepository(session)
            workflow = await repo.update_workflow(workflow_id, updates)
            if not workflow:
                raise ValueError(f"Workflow {workflow_id} not found")
            await session.commit()
            return {"id": workflow.id, "status": workflow.status.value}
        except Exception as e:
            await session.rollback()
            raise  # Let the outer function handle retries


@celery_app.task(
    name="Backend.celery_app.tasks.workflow_tasks.get_workflows_task",
    queue="workflow",
    priority=6
)
def get_workflows_task() -> List[Dict]:
    """Retrieve all workflows from the database."""
    return async_to_sync(_get_workflows_task)()


async def _get_workflows_task() -> List[Dict]:
    """Async implementation of get all workflows."""
    from Backend.data_layer.repositories.workflow_repository import WorkflowRepository
    from Backend.data_layer.database.connection import get_db

    async for session in get_db():
        repo = WorkflowRepository(session)
        workflows = await repo.get_all_workflows()
        return [{"id": w.id, "name": w.name, "status": w.status.value} for w in workflows]


@celery_app.task(
    name="Backend.celery_app.tasks.workflow_tasks.get_workflow_by_id_task",
    queue="workflow",
    priority=6
)
def get_workflow_by_id_task(workflow_id: int) -> Optional[Dict]:
    """Retrieve a workflow by ID from the database."""
    return async_to_sync(_get_workflow_by_id_task)(workflow_id)


async def _get_workflow_by_id_task(workflow_id: int) -> Optional[Dict]:
    """Async implementation of get workflow by ID."""
    from Backend.data_layer.repositories.workflow_repository import WorkflowRepository
    from Backend.data_layer.database.connection import get_db

    async for session in get_db():
        repo = WorkflowRepository(session)
        workflow = await repo.get_workflow(workflow_id)
        return workflow.to_dict() if workflow else None


@celery_app.task(
    name="tasks.workflow_tasks.collect_results",
    queue="workflow"
)
def collect_results(results: List[Dict]) -> Dict:
    """
    Collect and process the results of all workflow steps.
    """
    # FIX: Use StepStatus.COMPLETED instead of SUCCESS
    return {
        "status": StepStatus.COMPLETED if all(r.get("status") == StepStatus.COMPLETED for r in results) else StepStatus.FAILED,
        "steps": results
    }


@celery_app.task(
    name="tasks.workflow_tasks.process_workflow",
    queue="workflow",
    priority=7
)
def process_workflow(
    workflow_id: int,
    steps: List[Dict],
    user_id: int,
    context: Optional[Dict] = None
) -> Dict:
    """
    Process an entire workflow by executing its steps in sequence with execution tracking.
    """
    return async_to_sync(_process_workflow)(workflow_id, steps, user_id, context)


async def _process_workflow(
    workflow_id: int,
    steps: List[Dict],
    user_id: int,
    context: Optional[Dict] = None
) -> Dict:
    """Async implementation of workflow processing."""
    from Backend.data_layer.repositories.workflow_repository import WorkflowRepository
    from Backend.data_layer.database.connection import get_db

    async for session in get_db():
        try:
            repo = WorkflowRepository(session)

            # Create workflow execution record
            execution = await repo.create_workflow_execution(
                workflow_id=workflow_id,
                status=WorkflowStatus.ACTIVE
            )

            current_context = context or {}
            workflow_steps = []

            # Create step tasks with execution tracking
            for step in steps:
                step_task = execute_workflow_step.s(
                    workflow_id=workflow_id,
                    step_id=step["id"],
                    input_data={**step["input"], **current_context},
                    user_id=user_id,
                    execution_id=execution.id
                )
                workflow_steps.append(step_task)

            # Execute steps in parallel and collect results
            workflow_chain = group(workflow_steps) | collect_results.s()
            result = workflow_chain.apply_async()

            # Update workflow status
            await repo.update_workflow_status(workflow_id, WorkflowStatus.ACTIVE.value)
            await session.commit()

            return {
                "workflow_id": workflow_id,
                "execution_id": execution.id,
                "status": WorkflowStatus.ACTIVE.value,
                "task_id": result.id,
                "steps": [{"step_id": step["id"], "status": StepStatus.PENDING} for step in steps]
            }

        except Exception as e:
            await session.rollback()
            # Update workflow status to failed on error
            try:
                await repo.update_workflow_status(workflow_id, WorkflowStatus.FAILED.value)
                await session.commit()
            except Exception:
                await session.rollback()
            raise
