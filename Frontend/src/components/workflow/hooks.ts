import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import * as workflowApi from '@/components/workflow/api';
import { 
  CreateWorkflowRequest, 
  UpdateWorkflowRequest, 
  WorkflowStepRequest, 
  WorkflowTransitionRequest,
  WorkflowStep,
  WorkflowExecutionResponse
} from '@/components/workflow/types';
import { useMemo } from 'react';

// Query keys
export const workflowKeys = {
  all: ['workflows'] as const,
  lists: () => [...workflowKeys.all, 'list'] as const,
  list: () => [...workflowKeys.lists()] as const,
  details: () => [...workflowKeys.all, 'detail'] as const,
  detail: (id: string) => [...workflowKeys.details(), id] as const,
  executions: (workflowId: string) => [...workflowKeys.detail(workflowId), 'executions'] as const,
  steps: (workflowId: string) => [...workflowKeys.detail(workflowId), 'steps'] as const,
  step: (workflowId: string, stepId: string) => [...workflowKeys.steps(workflowId), stepId] as const,
  transitions: (workflowId: string) => [...workflowKeys.detail(workflowId), 'transitions'] as const,
};

// Workflow List Hook
export const useWorkflows = () => {
  return useQuery({
    queryKey: workflowKeys.list(),
    queryFn: () => workflowApi.fetchWorkflows(),
  });
};

// Workflow Detail Hook with Steps
export const useWorkflowDetail = (id: string) => {
  // First fetch the workflow details
  const workflowQuery = useQuery({
    queryKey: workflowKeys.detail(id),
    queryFn: () => workflowApi.getWorkflow(id),
    enabled: !!id,
  });

  // Then fetch the workflow steps
  const stepsQuery = useQuery({
    queryKey: workflowKeys.steps(id),
    queryFn: () => workflowApi.listSteps(id),
    enabled: !!id,
  });

  // Combine the data
  const data = useMemo(() => {
    if (workflowQuery.data && stepsQuery.data) {
      return {
        ...workflowQuery.data,
        steps: stepsQuery.data.steps || []
      };
    }
    return undefined;
  }, [workflowQuery.data, stepsQuery.data]);

  return {
    data,
    isLoading: workflowQuery.isLoading || stepsQuery.isLoading,
    error: workflowQuery.error || stepsQuery.error,
  };
};

// Step Operations Hooks
export const useWorkflowStep = (workflowId: string, stepId: string) => {
  return useQuery({
    queryKey: workflowKeys.step(workflowId, stepId),
    queryFn: () => workflowApi.getStep(workflowId, stepId),
    enabled: !!workflowId && !!stepId,
  });
};

// Create Workflow Hook
export const useCreateWorkflow = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (workflow: CreateWorkflowRequest) => workflowApi.createWorkflow(workflow),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.lists() });
    },
  });
};

// Update Workflow Hook
export const useUpdateWorkflow = (id: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (workflow: UpdateWorkflowRequest) => workflowApi.updateWorkflow(id, workflow),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: workflowKeys.lists() });
    },
  });
};

// Delete Workflow Hook
export const useDeleteWorkflow = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => workflowApi.deleteWorkflow(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.lists() });
    },
  });
};

// Execute Workflow Hook
export const useExecuteWorkflow = (workflowId: string) => {
  const queryClient = useQueryClient();

  return useMutation<WorkflowExecutionResponse, Error, void>({
    mutationFn: () => workflowApi.executeWorkflow(workflowId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.detail(workflowId) });
    },
  });
};

// Cancel Workflow Execution Hook
export const useCancelWorkflowExecution = (workflowId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (executionId: string) => workflowApi.cancelExecution(executionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.detail(workflowId) });
      queryClient.invalidateQueries({ queryKey: workflowKeys.executions(workflowId) });
    },
  });
};

// Workflow Executions Hook
export const useWorkflowExecutions = (workflowId: string, params?: {
  page?: number;
  page_size?: number;
  status?: string;
}) => {
  return useQuery({
    queryKey: [...workflowKeys.executions(workflowId), params],
    queryFn: () => workflowApi.listExecutions(workflowId, params),
    enabled: !!workflowId,
  });
};

// Workflow Steps Hooks
export const useWorkflowSteps = (workflowId: string) => {
  return useQuery({
    queryKey: workflowKeys.steps(workflowId),
    queryFn: () => workflowApi.listSteps(workflowId),
    enabled: !!workflowId,
  });
};

export const useCreateWorkflowStep = (workflowId: string) => {
  return useMutation({
    mutationFn: (step: WorkflowStepRequest) => workflowApi.createStep(workflowId, step),
  });
};

// Update Workflow Step Hook
export const useUpdateWorkflowStep = (workflowId: string) => {
  const queryClient = useQueryClient();

  return useMutation<
    WorkflowStep,
    Error,
    { stepId: string; step: Partial<WorkflowStepRequest> }
  >({
    mutationFn: ({ stepId, step }) => workflowApi.updateStep(workflowId, stepId, step),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.detail(workflowId) });
    },
  });
};

export const useDeleteWorkflowStep = (workflowId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (stepId: string) => workflowApi.deleteStep(workflowId, stepId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.detail(workflowId) });
    },
  });
};

export const useExecuteWorkflowStep = (workflowId: string) => {
  const queryClient = useQueryClient();

  return useMutation<WorkflowExecutionResponse, Error, string>({
    mutationFn: (stepId: string) => workflowApi.executeStep(workflowId, stepId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.detail(workflowId) });
    },
  });
};

// Workflow Transitions Hooks
export const useWorkflowTransitions = (workflowId: string) => {
  return useQuery({
    queryKey: workflowKeys.transitions(workflowId),
    queryFn: () => workflowApi.listTransitions(workflowId),
    enabled: !!workflowId,
  });
};

export const useCreateWorkflowTransition = (workflowId: string) => {
  return useMutation({
    mutationFn: (transition: WorkflowTransitionRequest) => 
      workflowApi.createTransition(workflowId, transition),
  });
};

export const useUpdateWorkflowTransition = (workflowId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ transitionId, transition }: { transitionId: string; transition: Partial<WorkflowTransitionRequest> }) => 
      workflowApi.updateTransition(workflowId, transitionId, transition),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.transitions(workflowId) });
      queryClient.invalidateQueries({ queryKey: workflowKeys.detail(workflowId) });
    },
  });
};

export const useDeleteWorkflowTransition = (workflowId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (transitionId: string) => workflowApi.deleteTransition(workflowId, transitionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.transitions(workflowId) });
      queryClient.invalidateQueries({ queryKey: workflowKeys.detail(workflowId) });
    },
  });
};

// Analysis and Optimization Hooks
export const useWorkflowAnalysis = (id: string) => {
  return useQuery({
    queryKey: [...workflowKeys.detail(id), 'analysis'],
    queryFn: () => workflowApi.analyzeWorkflow(id),
    enabled: !!id,
  });
};

export const useOptimizeWorkflow = (id: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => workflowApi.optimizeWorkflow(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.detail(id) });
    },
  });
}; 