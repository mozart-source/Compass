import { useState } from "react"
import { cn } from "@/lib/utils"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { 
  History, 
  Command, 
  Activity, 
  Brain,
  Plus, 
  ChevronRight,
  GitBranch,
  ArrowRight,
  Loader2
} from "lucide-react"
import { Button } from "../../ui/button"
import { useNavigate } from "react-router-dom"
import { WorkflowListItem, WorkflowStatus, WorkflowType } from "@/components/workflow/types"
import { Badge } from "../../ui/badge"
import { useWorkflows } from "@/components/workflow/hooks"
import WorkflowForm from "./WorkflowForm"
import { useCreateWorkflow } from "@/components/workflow/hooks"

const getStatusColor = (status: WorkflowStatus) => {
  switch (status) {
    case "active":
      return "bg-emerald-50 text-emerald-600 border-emerald-200 dark:bg-emerald-950/20 dark:text-emerald-400 dark:border-emerald-900/30"
    case "completed":
      return "bg-blue-50 text-blue-600 border-blue-200 dark:bg-blue-950/20 dark:text-blue-400 dark:border-blue-900/30"
    case "failed":
      return "bg-red-50 text-red-600 border-red-200 dark:bg-red-950/20 dark:text-red-400 dark:border-red-900/30"
    case "cancelled":
      return "bg-orange-50 text-orange-600 border-orange-200 dark:bg-orange-950/20 dark:text-orange-400 dark:border-orange-900/30"
    case "pending":
      return "bg-yellow-50 text-yellow-600 border-yellow-200 dark:bg-yellow-950/20 dark:text-yellow-400 dark:border-yellow-900/30"
    default:
      return "bg-gray-50 text-gray-600 border-gray-200 dark:bg-gray-950/20 dark:text-gray-400 dark:border-gray-900/30"
  }
}

const getWorkflowTypeIcon = (type: WorkflowType) => {
  switch (type) {
    case "sequential":
      return <ArrowRight className="h-4 w-4" />
    case "parallel":
      return <GitBranch className="h-4 w-4" />
    case "conditional":
      return <Activity className="h-4 w-4" />
    case "ai_driven":
      return <Brain className="h-4 w-4" />
    default:
      return <Command className="h-4 w-4" />
  }
}

const getWorkflowTypeStyle = (type: WorkflowType) => {
  switch (type) {
    case "sequential":
      return "text-blue-600 border-blue-200 bg-blue-50 dark:text-blue-400 dark:border-blue-950 dark:bg-blue-950/20"
    case "parallel":
      return "text-indigo-600 border-indigo-200 bg-indigo-50 dark:text-indigo-400 dark:border-indigo-950 dark:bg-indigo-950/20"
    case "conditional":
      return "text-amber-600 border-amber-200 bg-amber-50 dark:text-amber-400 dark:border-amber-950 dark:bg-amber-950/20"
    case "ai_driven":
      return "text-emerald-600 border-emerald-200 bg-emerald-50 dark:text-emerald-400 dark:border-emerald-950 dark:bg-emerald-950/20"
    default:
      return "text-gray-600 border-gray-200 bg-gray-50 dark:text-gray-400 dark:border-gray-950 dark:bg-gray-950/20"
  }
}

export default function WorkflowPage() {
  const navigate = useNavigate();
  const [showWorkflowForm, setShowWorkflowForm] = useState(false);

  const { data, isLoading, error } = useWorkflows();
  const createWorkflow = useCreateWorkflow();

  const handleWorkflowClick = (workflowId: string) => {
    navigate(`/workflow/${workflowId}`);
  };

  const handleNewWorkflow = () => {
    setShowWorkflowForm(true);
  };

  const handleWorkflowFormSubmit = async (formData: any) => {
    try {
      const organizationId = "838e655c-d07e-4a2e-8d06-ea6fef7e7b50";
      if (!organizationId) {
        console.error("No organization ID found");
        return;
      }

      await createWorkflow.mutateAsync({
        ...formData,
        organization_id: organizationId,
      });
      
      setShowWorkflowForm(false);
    } catch (error) {
      console.error("Failed to create workflow:", error);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-full text-destructive">
        Error loading workflows: {error.message}
      </div>
    );
  }

  const workflows = data || [];

  return (
    <div className="flex flex-1 flex-col gap-4 p-6 h-full">
      {/* Workflow Label */}
      <p className="text-xs uppercase text-muted-foreground tracking-wider">workflow</p>
      <div className="flex justify-start">
          <h1 className="text-2xl font-bold tracking-tight leading-none">Workflow Automation</h1>
        <div className="ml-auto">
          <Button 
            variant="outline" 
            size="sm" 
            className="gap-2"
            onClick={handleNewWorkflow}
          >
            <Plus className="h-4 w-4" />
            New Workflow
          </Button>
        </div>
      </div>

      {/* Workflow Cards */}
      <div className="space-y-4">
        {workflows.length === 0 ? (
          <div className="text-center text-muted-foreground py-8">
            No workflows found. Create a new workflow to get started.
          </div>
        ) : (
          workflows.map((workflow: WorkflowListItem) => (
            <Card 
              key={workflow.id} 
              className={cn(
                "cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors"
              )}
              onClick={() => handleWorkflowClick(workflow.id)}
            >
              <CardHeader>
                <div className="flex justify-between items-center">
                  <div className="flex items-center gap-3">
                    <CardTitle className="text-lg">{workflow.name}</CardTitle>
                    <Badge variant="outline" className={cn(
                      "text-xs",
                      getStatusColor(workflow.status)
                    )}>
                      {workflow.status}
                    </Badge>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge variant="outline" className={cn(
                      "gap-1.5",
                      getWorkflowTypeStyle(workflow.workflowType)
                    )}>
                      {getWorkflowTypeIcon(workflow.workflowType)}
                      {workflow.workflowType}
                    </Badge>
                    <ChevronRight className="h-5 w-5 text-muted-foreground" />
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground mb-2">{workflow.description}</p>
                <div className="flex items-center gap-4">
                  <p className="text-xs text-muted-foreground flex items-center gap-1">
                    <History className="h-3.5 w-3.5" />
                    Last Run: {workflow.lastExecutedAt ? new Date(workflow.lastExecutedAt).toLocaleString() : 'Never'}
                  </p>
                  {workflow.tags && workflow.tags.length > 0 && (
                    <div className="flex items-center gap-1">
                      {workflow.tags.map((tag: string, index: number) => (
                        <Badge key={index} variant="secondary" className="text-xs">
                          {tag}
                        </Badge>
                      ))}
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          ))
        )}
      </div>

      {/* Workflow Form Modal */}
      {showWorkflowForm && (
        <WorkflowForm
          onClose={() => setShowWorkflowForm(false)}
          onSubmit={handleWorkflowFormSubmit}
        />
      )}
    </div>
  )
}
