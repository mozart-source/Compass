import axios from 'axios';
import { 
  WorkflowDetail, 
  WorkflowListItem, 
  CreateWorkflowRequest, 
  UpdateWorkflowRequest,
  WorkflowStepRequest,
  WorkflowTransitionRequest,
  WorkflowExecutionResponse,
  WorkflowAnalysisResponse
} from '@/components/workflow/types';
import { getApiUrls } from '@/config';

const { GO_API_URL } = getApiUrls();
const API_BASE_URL = `${GO_API_URL}/workflows`;

axios.defaults.headers.common['X-Organization-ID'] = '838e655c-d07e-4a2e-8d06-ea6fef7e7b50';

// Configure axios defaults
axios.defaults.timeout = 30000; // 30 seconds

// Workflow operations
export const fetchWorkflows = async (): Promise<WorkflowListItem[]> => {
  const response = await axios.get<{ data: { workflows: WorkflowListItem[] } }>(`${API_BASE_URL}`);
  return response.data.data.workflows;
};

export const getWorkflow = async (id: string): Promise<WorkflowDetail> => {
  const response = await axios.get<{ data: WorkflowDetail }>(`${API_BASE_URL}/${id}`);
  return response.data.data;
};

export const createWorkflow = async (workflow: CreateWorkflowRequest): Promise<WorkflowDetail> => {
  const response = await axios.post<{ data: WorkflowDetail }>(`${API_BASE_URL}`, workflow);
  return response.data.data;
};

export const updateWorkflow = async (id: string, workflow: UpdateWorkflowRequest): Promise<WorkflowDetail> => {
  const response = await axios.put<{ data: WorkflowDetail }>(`${API_BASE_URL}/${id}`, workflow);
  return response.data.data;
};

export const deleteWorkflow = async (id: string): Promise<void> => {
  await axios.delete(`${API_BASE_URL}/${id}`);
};

// Workflow execution operations
export const executeWorkflow = async (id: string): Promise<WorkflowExecutionResponse> => {
  const response = await axios.post<{ data: WorkflowExecutionResponse }>(`${API_BASE_URL}/${id}/execute`);
  return response.data.data;
};

export const cancelExecution = async (executionId: string): Promise<void> => {
  await axios.post(`${API_BASE_URL}/executions/${executionId}/cancel`);
};

export const getExecution = async (executionId: string): Promise<WorkflowExecutionResponse> => {
  const response = await axios.get<{ data: WorkflowExecutionResponse }>(`${API_BASE_URL}/executions/${executionId}`);
  return response.data.data;
};

export const listExecutions = async (workflowId: string, params?: {
  page?: number;
  page_size?: number;
  status?: string;
}): Promise<{ executions: WorkflowExecutionResponse[]; total: number }> => {
  const response = await axios.get<{ data: { executions: WorkflowExecutionResponse[]; total: number } }>(`${API_BASE_URL}/${workflowId}/executions`, {
    params: params || {}
  });
  return response.data.data;
};

// Step operations
export const listSteps = async (workflowId: string) => {
  const response = await axios.get(`${API_BASE_URL}/${workflowId}/steps`);
  return response.data;
};

export const getStep = async (workflowId: string, stepId: string) => {
  const response = await axios.get(`${API_BASE_URL}/${workflowId}/steps/${stepId}`);
  return response.data;
};

export const createStep = async (workflowId: string, step: WorkflowStepRequest) => {
  const response = await axios.post(`${API_BASE_URL}/${workflowId}/steps`, step);
  return response.data;
};

export const updateStep = async (workflowId: string, stepId: string, step: Partial<WorkflowStepRequest>) => {
  const response = await axios.put(`${API_BASE_URL}/${workflowId}/steps/${stepId}`, step);
  return response.data;
};

export const deleteStep = async (workflowId: string, stepId: string): Promise<void> => {
  await axios.delete(`${API_BASE_URL}/${workflowId}/steps/${stepId}`);
};

export const executeStep = async (workflowId: string, stepId: string) => {
  const response = await axios.post(`${API_BASE_URL}/${workflowId}/steps/${stepId}/execute`);
  return response.data;
};

// Transition operations
export const listTransitions = async (workflowId: string) => {
  const response = await axios.get(`${API_BASE_URL}/${workflowId}/transitions`);
  return response.data;
};

export const getTransition = async (workflowId: string, transitionId: string) => {
  const response = await axios.get(`${API_BASE_URL}/${workflowId}/transitions/${transitionId}`);
  return response.data;
};

export const createTransition = async (workflowId: string, transition: WorkflowTransitionRequest) => {
  const response = await axios.post(`${API_BASE_URL}/${workflowId}/transitions`, transition);
  return response.data;
};

export const updateTransition = async (
  workflowId: string, 
  transitionId: string, 
  transition: Partial<WorkflowTransitionRequest>
) => {
  const response = await axios.put(`${API_BASE_URL}/${workflowId}/transitions/${transitionId}`, transition);
  return response.data;
};

export const deleteTransition = async (workflowId: string, transitionId: string): Promise<void> => {
  await axios.delete(`${API_BASE_URL}/${workflowId}/transitions/${transitionId}`);
};

// Analysis and optimization operations
export const analyzeWorkflow = async (id: string): Promise<WorkflowAnalysisResponse> => {
  const response = await axios.post<{ data: WorkflowAnalysisResponse }>(`${API_BASE_URL}/${id}/analyze`);
  return response.data.data;
};

export const optimizeWorkflow = async (id: string): Promise<WorkflowAnalysisResponse> => {
  const response = await axios.post<{ data: WorkflowAnalysisResponse }>(`${API_BASE_URL}/${id}/optimize`);
  return response.data.data;
}; 