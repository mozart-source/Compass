import { useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { cn } from "@/lib/utils";
import { 
  ChevronLeft, 
  CheckCircle2, 
  CircleDot, 
  AlertCircle, 
  Clock,
  CheckIcon,
  ArrowRightCircle,
  Loader2,
  Brain,
  Plus
} from "lucide-react";
import { Button } from "../../ui/button";
import { Badge } from "@/components/ui/badge";
import { motion } from "framer-motion";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { WorkflowDetail, WorkflowStep, StepStatus, WorkflowStepRequest } from "@/components/workflow/types";
import { calculateLineAnimations } from "@/utils/workflowAnimation";
import ParallelWorkflowDetailPage from "./ParallelWorkflowDetail";
import { useWorkflowDetail, useUpdateWorkflowStep, useExecuteWorkflow, useExecuteWorkflowStep, useCreateWorkflowStep, useCreateWorkflowTransition } from "@/components/workflow/hooks";
import { UseQueryResult, useQueryClient } from "@tanstack/react-query";
import WorkflowStepForm from "./WorkflowStepForm";
import WorkflowStepEditForm from "./WorkflowStepEditForm";

interface WorkflowDetailProps {
  darkMode?: boolean;
}

export default function WorkflowDetailPage({ darkMode = false }: WorkflowDetailProps) {
  const { id } = useParams<{ id: string }>();
  const { data: workflow, isLoading, error } = useWorkflowDetail(id!) as UseQueryResult<WorkflowDetail, Error>;

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (error || !workflow) {
    return (
      <div className="flex flex-1 flex-col p-6 h-[calc(100vh-32px)] overflow-hidden">
        <p className="text-destructive">Error loading workflow</p>
      </div>
    );
  }

  // Render different workflow visualizations based on workflow type
  switch (workflow.workflowType) {
    case "parallel":
      return <ParallelWorkflowDetailPage darkMode={darkMode} mockData={workflow} />;
    case "sequential":
    default:
      return <SequentialWorkflowDetail workflow={workflow} workflowId={id!} />;
  }
}

// Sequential workflow visualization component
interface SequentialWorkflowDetailProps {
  workflow: WorkflowDetail;
  workflowId: string;
}

function SequentialWorkflowDetail({ workflow, workflowId }: SequentialWorkflowDetailProps) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [activeStep, setActiveStep] = useState<string | null>(null);
  const [completedSteps, setCompletedSteps] = useState<string[]>([]);
  const [expandedStep, setExpandedStep] = useState<string | null>(null);
  const [lastCompletedStep, setLastCompletedStep] = useState<string | null>(null);
  const [isAddingStep, setIsAddingStep] = useState(false);
  const [editingStep, setEditingStep] = useState<WorkflowStep | null>(null);

  const updateStep = useUpdateWorkflowStep(workflowId);
  const executeWorkflow = useExecuteWorkflow(workflowId);
  const executeStep = useExecuteWorkflowStep(workflowId);
  const createStep = useCreateWorkflowStep(workflowId);
  const createTransition = useCreateWorkflowTransition(workflowId);

  // Initialize from workflow data
  useState(() => {
    if (workflow.steps.length > 0) {
      // Set first step as active by default
      setActiveStep(workflow.steps[0].id);
      
      // Set completed steps based on the workflow data
      const completed = workflow.steps
        .filter(step => step.status === "completed")
        .map(step => step.id);
        
      setCompletedSteps(completed);
      
      if (completed.length > 0) {
        setLastCompletedStep(completed[completed.length - 1]);
      }
    }
  });

  const handleBack = () => {
    navigate(-1); // Go back to the previous page
  };

  const handleAddNewStep = async (stepData: Partial<WorkflowStepRequest>) => {
    try {
      const lastStep = workflow.steps.length > 0 ? workflow.steps[workflow.steps.length - 1] : null;
      const newStepOrder = lastStep ? lastStep.step_order + 1 : 1;
      
      const newStep = await createStep.mutateAsync({
        ...stepData,
        step_order: newStepOrder,
        name: stepData.name!,
        step_type: stepData.step_type!,
        description: stepData.description!,
      });

      if (lastStep) {
        await createTransition.mutateAsync({
          from_step_id: lastStep.id,
          to_step_id: newStep.id,
        });
      }

      await queryClient.invalidateQueries({
        queryKey: ["workflows", "detail", workflowId],
      });

      setIsAddingStep(false);
    } catch (error) {
      console.error("Failed to add new step:", error);
    }
  };

  const handleUpdateStep = async (stepData: Partial<WorkflowStepRequest>) => {
    if (!editingStep) return;
    try {
      await updateStep.mutateAsync({
        stepId: editingStep.id,
        step: stepData,
      });
      setEditingStep(null);
    } catch (error) {
      console.error("Failed to update step:", error);
    }
  };

  const handleAutoRun = async () => {
    try {
      await executeWorkflow.mutateAsync();
    } catch (error) {
      console.error('Failed to execute workflow:', error);
    }
  };

  const handleStepClick = async (stepId: string) => {
    const step = workflow.steps.find(s => s.id === stepId);
    if (!step) return;

    // Toggle expanded state
    setExpandedStep(expandedStep === stepId ? null : stepId);
    setActiveStep(stepId);

    // If step is auto-advance and not completed, execute it
    if (step.autoAdvance && step.status !== "completed") {
      try {
        await executeStep.mutateAsync(stepId);
      } catch (error) {
        console.error('Failed to execute step:', error);
      }
    }
  };

  const handleToggleStepCompletion = async (stepId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    const step = workflow.steps.find(s => s.id === stepId);
    if (!step || !workflowId) return;

    const newStatus = step.status === "completed" ? "pending" : "completed";
    
    try {
      await updateStep.mutateAsync({
        stepId,
        step: { 
          status: newStatus as StepStatus,
          step_type: step.step_type
        }
      });

      const newCompletedSteps = newStatus === "completed"
        ? [...completedSteps, stepId]
        : completedSteps.filter(id => id !== stepId);
        
      setCompletedSteps(newCompletedSteps);
      
      if (newStatus === "completed") {
        setLastCompletedStep(stepId);
        
        // If auto-advance is enabled and there's a next step, activate it
        const currentIndex = workflow.steps.findIndex(s => s.id === stepId);
        if (step.autoAdvance && currentIndex < workflow.steps.length - 1) {
          const nextStep = workflow.steps[currentIndex + 1];
          setActiveStep(nextStep.id);
          setExpandedStep(nextStep.id);
        }
      } else if (stepId === lastCompletedStep) {
        const remainingCompleted = newCompletedSteps.filter(id => id !== stepId);
        setLastCompletedStep(remainingCompleted.length > 0 ? remainingCompleted[remainingCompleted.length - 1] : null);
      }
    } catch (error) {
      console.error('Failed to update step status:', error);
    }
  };

  const getStepIcon = (type: string, isCompleted: boolean = false) => {
    if (isCompleted) {
      return <CheckCircle2 className="h-6 w-6 text-white" />;
    }
    
    switch (type) {
      case "start":
        return <CircleDot className="h-6 w-6 text-emerald-500" />;
      case "process":
        return <ArrowRightCircle className="h-6 w-6 text-primary" />;
      case "decision":
        return <AlertCircle className="h-6 w-6 text-amber-500" />;
      case "end":
        return <CheckCircle2 className="h-6 w-6 text-destructive" />;
      case "manual":
        return <CircleDot className="h-6 w-6 text-blue-500" />;
      case "automated":
        return <ArrowRightCircle className="h-6 w-6 text-emerald-500" />;
      case "approval":
        return <CheckCircle2 className="h-6 w-6 text-amber-500" />;
      case "notification":
        return <AlertCircle className="h-6 w-6 text-indigo-500" />;
      case "integration":
        return <ArrowRightCircle className="h-6 w-6 text-violet-500" />;
      case "ai_task":
        return <Brain className="h-6 w-6 text-cyan-500" />;
      default:
        return <CircleDot className="h-6 w-6" />;
    }
  };

  const getStepBadgeStyles = (type: string) => {
    switch (type) {
      case "start":
        return "text-emerald-600 border-emerald-200 bg-emerald-50 dark:text-emerald-400 dark:border-emerald-950 dark:bg-emerald-950 dark:bg-opacity-20";
      case "process":
        return "text-primary border-primary/20 bg-primary/10 dark:text-primary dark:border-primary/20 dark:bg-primary/10";
      case "decision":
        return "text-amber-600 border-amber-200 bg-amber-50 dark:text-amber-400 dark:border-amber-950 dark:bg-amber-950 dark:bg-opacity-20";
      case "end":
        return "text-destructive border-destructive/20 bg-destructive/10 dark:text-destructive dark:border-destructive/20 dark:bg-destructive/10";
      case "manual":
        return "text-blue-600 border-blue-200 bg-blue-50 dark:text-blue-400 dark:border-blue-950 dark:bg-blue-950 dark:bg-opacity-20";
      case "automated":
        return "text-emerald-600 border-emerald-200 bg-emerald-50 dark:text-emerald-400 dark:border-emerald-950 dark:bg-emerald-950 dark:bg-opacity-20";
      case "approval":
        return "text-amber-600 border-amber-200 bg-amber-50 dark:text-amber-400 dark:border-amber-950 dark:bg-amber-950 dark:bg-opacity-20";
      case "notification":
        return "text-indigo-600 border-indigo-200 bg-indigo-50 dark:text-indigo-400 dark:border-indigo-950 dark:bg-indigo-950 dark:bg-opacity-20";
      case "integration":
        return "text-violet-600 border-violet-200 bg-violet-50 dark:text-violet-400 dark:border-violet-950 dark:bg-violet-950 dark:bg-opacity-20";
      case "ai_task":
        return "text-cyan-600 border-cyan-200 bg-cyan-50 dark:text-cyan-400 dark:border-cyan-950 dark:bg-cyan-950 dark:bg-opacity-20";
      default:
        return "text-muted-foreground border-border bg-muted dark:text-muted-foreground dark:border-border dark:bg-muted";
    }
  };

  const getStepRingStyle = (type: string) => {
    switch (type) {
      case "start":
        return "ring-emerald-500";
      case "process":
        return "ring-primary";
      case "decision":
        return "ring-amber-500";
      case "end":
        return "ring-destructive";
      default:
        return "ring-muted-foreground";
    }
  };

  // Calculate progress percentage
  const progressPercentage = workflow.steps.length > 0
    ? (completedSteps.length / workflow.steps.length) * 100
    : 0;

  // Group steps into rows of 3
  const groupedSteps = workflow.steps.reduce<WorkflowStep[][]>((acc, step, index) => {
    const rowIndex = Math.floor(index / 3);
    if (!acc[rowIndex]) {
      acc[rowIndex] = [];
    }
    acc[rowIndex].push(step);
    return acc;
  }, []);
  
  // Calculate animations for all lines
  const { horizontalLines, verticalLines } = calculateLineAnimations(groupedSteps, completedSteps);

  return (
    <div className="flex flex-1 flex-col gap-4 p-6 h-[calc(100vh-32px)] overflow-hidden">
      {/* Header */}
      <div className="flex justify-start">
        <div className="flex items-center gap-2">
          <Button 
            variant="ghost" 
            size="sm" 
            onClick={handleBack}
            className="p-0 h-8 w-8 mr-1"
          >
            <ChevronLeft className="h-5 w-5" />
          </Button>
          <h2 className="text-3xl font-bold tracking-tight">{workflow.name}</h2>
        </div>
        {/* Workflow Status Bar - simplified to a single line */}
        <div className="flex">
          <div className="flex items-center gap-2">
            <Badge variant="outline" className="text-xs">
              {workflow.steps.length} steps
            </Badge>
            <Badge 
              variant="outline" 
              className={cn(
                "text-xs",
                completedSteps.length === workflow.steps.length 
                  ? "bg-emerald-50 text-emerald-600 border-emerald-200 dark:bg-emerald-950/20 dark:text-emerald-400 dark:border-emerald-900/30"
                  : "bg-amber-50 text-amber-600 border-amber-200 dark:bg-amber-950/20 dark:text-amber-400 dark:border-amber-900/30"
              )}
            >
              {completedSteps.length === workflow.steps.length ? "Complete" : "In Progress"}
            </Badge>
            <span className="text-xs text-muted-foreground flex items-center">
              <Clock className="h-3 w-3 mr-1 inline" />
              {workflow.lastExecutedAt ? new Date(workflow.lastExecutedAt).toLocaleDateString() : 'Never'}
            </span>
          </div>
        </div>

        <div className="flex items-center gap-2 ml-auto">
          <Button 
            variant="outline" 
            size="sm" 
            className="gap-1.5 h-8"
            onClick={() => setIsAddingStep(true)}
          >
            <Plus className="h-3.5 w-3.5" />
            New Step
          </Button>
          <Button 
            variant="outline" 
            size="sm" 
            className="gap-1.5 h-8"
            onClick={handleAutoRun}
            disabled={executeWorkflow.isPending}
          >
            {executeWorkflow.isPending ? (
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
            ) : (
              <CheckIcon className="h-3.5 w-3.5" />
            )}
            Execute Workflow
          </Button>
        </div>
      </div>

      {/* Main content container */}
      <div className="flex-1 flex flex-col min-h-0">

        {/* Workflow Steps Container */}
        <div className="flex-1 bg-background border border-border rounded-xl overflow-hidden flex flex-col min-h-0">
          <div className="p-4 border-b border-border flex items-center justify-between bg-card">
            <h3 className="font-medium">Sequential Workflow</h3>
            <div className="flex items-center text-sm">
              <span className="mr-2 text-muted-foreground">{Math.round(progressPercentage)}% Completed</span>
              <div className="w-16 h-1.5 bg-muted rounded-full overflow-hidden">
                <motion.div 
                  className="h-full bg-primary rounded-full"
                  initial={{ width: 0 }}
                  animate={{ width: `${progressPercentage}%` }}
                  transition={{ duration: 0.5, ease: "easeOut" }}
                />
              </div>
            </div>
          </div>
          
          <div className="flex-1 overflow-auto p-4 min-h-0">
            {groupedSteps.map((rowSteps, rowIndex) => {
              const isRtl = rowIndex % 2 !== 0; // Every odd row is right-to-left
              // Don't reverse the steps array, we'll handle the visual order with flex-row-reverse
              const sortedSteps = rowSteps;
              
              return (
                <div key={`row-${rowIndex}`} className={`mb-16 relative ${rowIndex === 0 ? 'pt-4' : ''}`}>
                  {/* Background horizontal line for the row */}
                  <div 
                    className="absolute h-1.5 bg-border dark:bg-muted rounded-full z-0 overflow-hidden"
                    style={{
                      top: 'calc(50% + 12px)', // Adjusted position
                      left: '15%', // Pushed inward from left side
                      right: '15%', // Pushed inward from right side
                      transform: 'translateY(-50%)'
                    }}
                  >
                    {/* Animated horizontal line fill based on completion */}
                    <motion.div 
                      className="h-full bg-primary rounded-full"
                      initial={{ width: 0 }}
                      animate={{ width: `${horizontalLines[rowIndex]}%` }}
                      transition={{ 
                        duration: 0.5, 
                        ease: "easeOut",
                        // If there's a previous row, delay this row's animation slightly to create a sequence effect
                        delay: rowIndex > 0 ? 0.05 * rowIndex : 0
                      }}
                      style={{
                        transformOrigin: isRtl ? 'right' : 'left',
                        marginLeft: isRtl ? 'auto' : '0'
                      }}
                    />
                  </div>
                  
                  {/* Vertical connecting line to previous row */}
                  {rowIndex > 0 && (
                    <div 
                      className="absolute w-1.5 bg-border dark:bg-muted rounded-full z-0 overflow-hidden"
                      style={{
                        top: '-70px', // Connect to previous row
                        height: '70px',
                        // Position based on the zigzag pattern - correctly align connection points
                        // Apply different adjustments based on the direction:
                        // - For LTR to RTL transitions (row 1 to 2): move right (+5px)
                        // - For RTL to LTR transitions (row 2 to 3): move left (-5px)
                        left: `calc(${(rowIndex-1) % 2 === 0 ? '85%' : '15%'} ${(rowIndex-1) % 2 === 0 ? '+' : '-'} 17px)`,
                        transform: 'translateX(-50%)'
                      }}
                    >
                      {/* Animated vertical line fill based on completion */}
                      <motion.div 
                        className="w-full bg-primary rounded-full"
                        initial={{ height: 0 }}
                        animate={{ height: `${verticalLines[rowIndex-1]}%` }}
                        transition={{ 
                          duration: 0.4, 
                          ease: "easeOut",
                          // Time the vertical animation to happen after the previous row's horizontal animation
                          // but before the current row's horizontal animation
                          delay: 0.05 * (rowIndex - 0.5)
                        }}
                        style={{ transformOrigin: 'top' }}
                      />
                    </div>
                  )}
                  
                  <div className={`flex items-center justify-between px-4 relative ${isRtl ? 'flex-row-reverse' : 'flex-row'}`}>
                    {sortedSteps.map((step, colIndex) => {
                      // Use simple index calculation since we're handling direction with CSS
                      const actualIndex = colIndex;
                      const stepIndex = rowIndex * 3 + actualIndex;
                      const isCompleted = completedSteps.includes(step.id);
                      const isExpanded = expandedStep === step.id;
                      
                      return (
                        <motion.div 
                          key={step.id} 
                          className="flex flex-col items-center relative z-10"
                          initial={{ opacity: 0, y: 20 }}
                          animate={{ opacity: 1, y: 0 }}
                          transition={{ duration: 0.3, delay: stepIndex * 0.1 }}
                          style={{ width: 'calc(27.33% - 1rem)' }}
                        >
                          {/* Step Node */}
                          <motion.div 
                            className={cn(
                              "relative w-full p-5 rounded-xl shadow-md transition-all duration-200 mt-4",
                              "bg-card border border-border",
                              activeStep === step.id && `ring-2 ${getStepRingStyle(step.step_type)} ring-opacity-70`,
                              isCompleted && "border-primary/30 dark:border-primary/30",
                              "cursor-pointer group z-10"
                            )}
                            layout
                            onClick={() => handleStepClick(step.id)}
                            whileHover={{ scale: 1.02, y: -2 }}
                            transition={{ 
                              layout: { duration: 0.2 },
                              scale: { duration: 0.1 }
                            }}
                          >
                            {/* Step Icon - Improved positioning and sizing */}
                            <motion.div
                              className={cn(
                                "absolute top-0 rounded-full p-2.5 shadow-md border border-border",
                                "bg-background group-hover:scale-110 transition-transform",
                                isCompleted ? "bg-emerald-600" : ""
                              )}
                              onClick={(e) => handleToggleStepCompletion(step.id, e)}
                              style={{ 
                                left: '50%',
                                transform: 'translate(-50%, -50%)',
                                zIndex: 30
                              }}
                            >
                              {getStepIcon(step.step_type, isCompleted)}
                            </motion.div>
                            
                            <div className="w-full pt-4">
                              <div className="flex justify-between items-start">
                                <h3 className={cn(
                                  "font-medium text-base text-card-foreground",
                                  isCompleted && "text-primary dark:text-primary"
                                )}>{step.name}</h3>
                                <div className="flex items-center gap-1">
                                  <Badge variant="outline" className={cn(
                                    "text-xs capitalize",
                                    getStepBadgeStyles(step.step_type)
                                  )}>
                                    {step.step_type}
                                  </Badge>
                                  <Button
                                    size="sm"
                                    variant="outline"
                                    className="h-7 px-2 text-xs"
                                    onClick={(e) => {
                                      e.stopPropagation();
                                      setEditingStep(step);
                                    }}
                                  >
                                    Edit
                                  </Button>
                                </div>
                              </div>
                              <p className="text-sm text-muted-foreground mt-1.5">{step.description}</p>
                              
                              {/* Expanded Actions - Only show when expanded */}
                              {isExpanded && (
                                <motion.div 
                                  className="mt-4 pt-3 border-t border-border flex justify-end gap-2"
                                  initial={{ opacity: 0, height: 0 }}
                                  animate={{ opacity: 1, height: "auto" }}
                                  exit={{ opacity: 0, height: 0 }}
                                >
                                  <TooltipProvider>
                                    <Tooltip>
                                      <TooltipTrigger asChild>
                                        <Button 
                                          size="sm" 
                                          variant="outline"
                                          className="h-8 px-2.5 text-xs"
                                          onClick={(e) => {
                                            e.stopPropagation();
                                            handleToggleStepCompletion(step.id, e);
                                          }}
                                          disabled={updateStep.isPending}
                                        >
                                          {updateStep.isPending ? (
                                            <Loader2 className="h-4 w-4 animate-spin" />
                                          ) : (
                                            step.status === "completed" ? "Mark Incomplete" : "Mark Complete"
                                          )}
                                        </Button>
                                      </TooltipTrigger>
                                      <TooltipContent>
                                        <p>Toggle step completion status</p>
                                      </TooltipContent>
                                    </Tooltip>
                                  </TooltipProvider>
                                </motion.div>
                              )}
                            </div>
                          </motion.div>
                        </motion.div>
                      );
                    })}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>
      {isAddingStep && (
        <WorkflowStepForm
          onClose={() => setIsAddingStep(false)}
          onSubmit={handleAddNewStep}
          isLoading={createStep.isPending || createTransition.isPending}
        />
      )}
      {editingStep && (
        <WorkflowStepEditForm
          step={editingStep}
          onClose={() => setEditingStep(null)}
          onSubmit={handleUpdateStep}
          isLoading={updateStep.isPending}
        />
      )}
    </div>
  );
}