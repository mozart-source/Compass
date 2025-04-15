export type WorkflowStatus = 
  | "pending"
  | "active"
  | "paused"
  | "completed"
  | "failed"
  | "cancelled"
  | "archived"
  | "under_review"
  | "optimizing";

export type WorkflowType = 
  | "sequential"
  | "parallel"
  | "conditional"
  | "ai_driven"
  | "hybrid";

export type StepType = 
  | "manual"
  | "automated"
  | "approval"
  | "notification"
  | "integration"
  | "decision"
  | "ai_task";

export type StepStatus = 
  | "pending"
  | "active"
  | "completed"
  | "failed"
  | "cancelled"
  | "skipped";

export interface WorkflowStep {
  id: string;
  workflowId: string;
  name: string;
  description: string;
  step_type: StepType;
  step_order: number;
  status: StepStatus;
  config?: Record<string, unknown>;
  conditions?: Record<string, unknown>;
  timeout?: number;
  retryConfig?: Record<string, unknown>;
  isRequired: boolean;
  autoAdvance: boolean;
  canRevert: boolean;
  dependencies?: string[];
  version: string;
  previousVersionId?: string;
  averageExecutionTime: number;
  successRate: number;
  lastExecutionResult?: Record<string, unknown>;
  assignedTo?: string;
  notificationConfig?: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
}

export interface WorkflowTransition {
  id: string;
  from_step_id: string;
  to_step_id: string;
  condition?: string;
  workflowId: string;
  conditions?: Record<string, unknown>;
  triggers?: Record<string, unknown>;
}

export interface WorkflowDetail {
  id: string;
  name: string;
  description: string;
  workflowType: WorkflowType;
  createdBy: string;
  organizationId: string;
  status: WorkflowStatus;
  config?: Record<string, unknown>;
  workflowMetadata?: Record<string, unknown>;
  version: string;
  tags: string[];
  
  // AI Integration
  aiEnabled: boolean;
  aiConfidenceThreshold?: number;
  aiOverrideRules?: Record<string, unknown>;
  aiLearningData?: Record<string, unknown>;

  // Performance Metrics
  averageCompletionTime?: number;
  successRate?: number;
  optimizationScore?: number;
  bottleneckAnalysis?: Record<string, unknown>;

  // Time Management
  estimatedDuration?: number;
  actualDuration?: number;
  scheduleConstraints?: Record<string, unknown>;
  deadline?: string;

  // Error Handling
  errorHandlingConfig?: Record<string, unknown>;
  retryPolicy?: Record<string, unknown>;
  fallbackSteps?: Record<string, unknown>;

  // Audit & Compliance
  complianceRules?: Record<string, unknown>;
  auditTrail?: Record<string, unknown>;
  accessControl?: Record<string, unknown>;

  // Timestamps
  createdAt: string;
  updatedAt: string;
  lastExecutedAt?: string;
  nextScheduledRun?: string;

  // Relations
  steps: WorkflowStep[];
  transitions: WorkflowTransition[];
}

export interface WorkflowListItem {
  id: string;
  name: string;
  description: string;
  workflowType: WorkflowType;
  createdBy: string;
  organizationId: string;
  status: WorkflowStatus;
  config: Record<string, unknown>;
  workflowMetadata: {
    version: string;
    createdAt: string;
    creatorId: string;
  };
  version: string;
  tags: string[] | null;
  aiEnabled: boolean;
  aiConfidenceThreshold: number;
  aiOverrideRules: Record<string, unknown>;
  aiLearningData: Record<string, unknown>;
  averageCompletionTime: number;
  successRate: number;
  optimizationScore: number;
  bottleneckAnalysis: Record<string, unknown>;
  estimatedDuration: number | null;
  actualDuration: number | null;
  scheduleConstraints: Record<string, unknown>;
  deadline: string | null;
  errorHandlingConfig: Record<string, unknown>;
  retryPolicy: Record<string, unknown>;
  fallbackSteps: Record<string, unknown>;
  complianceRules: Record<string, unknown>;
  auditTrail: Record<string, unknown>;
  accessControl: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
  lastExecutedAt: string | null;
  nextScheduledRun: string | null;
}

export interface CreateWorkflowRequest {
  name: string;
  description: string;
  workflow_type: WorkflowType;
  organization_id: string;
  config?: Record<string, unknown>;
  ai_enabled?: boolean;
  tags?: string[];
  estimated_duration?: number;
  deadline?: string;
}

export interface UpdateWorkflowRequest {
  name?: string;
  description?: string;
  status?: WorkflowStatus;
  config?: Record<string, unknown>;
  aiEnabled?: boolean;
  tags?: string[];
  estimatedDuration?: number;
  deadline?: string;
}

export interface WorkflowStepRequest {
  name: string;
  description: string;
  step_type: StepType;
  step_order: number;
  status?: StepStatus;
  config?: Record<string, unknown>;
  conditions?: Record<string, unknown>;
  timeout?: number;
  isRequired?: boolean;
  autoAdvance?: boolean;
  canRevert?: boolean;
  dependencies?: string[];
  assignedTo?: string;
  notificationConfig?: Record<string, unknown>;
}

export interface WorkflowTransitionRequest {
  from_step_id: string;
  to_step_id: string;
  conditions?: Record<string, unknown>;
  triggers?: Record<string, unknown>;
}

export interface WorkflowExecutionResponse {
  id: string;
  workflowId: string;
  status: WorkflowStatus;
  startedAt: string;
  completedAt?: string;
  error?: string;
  result?: Record<string, unknown>;
  metadata?: Record<string, unknown>;
}

export interface WorkflowAnalysisResponse {
  workflowId: string;
  metrics: {
    averageCompletionTime: number;
    successRate: number;
    bottlenecks: Array<{
      stepId: string;
      name: string;
      averageExecutionTime: number;
      recommendation: string;
    }>;
  };
  recommendations: Array<{
    type: string;
    description: string;
    priority: 'high' | 'medium' | 'low';
  }>;
} 