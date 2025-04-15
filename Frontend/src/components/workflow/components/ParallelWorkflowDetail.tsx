import { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { cn } from "@/lib/utils";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { 
  ChevronLeft,
  CheckCircle2,
  CircleDot,
  AlertCircle,
  ArrowRightCircle,
  CheckIcon,
  GitMerge,
  GitBranch,
  Clock,
  MoreHorizontal
} from "lucide-react";
import { Button } from "../../ui/button";
import { Badge } from "@/components/ui/badge";
import { motion } from "framer-motion";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { WorkflowDetail, WorkflowStep } from "@/components/workflow/types";
import { calculateParallelAnimations } from "@/utils/parallelWorkflowAnimation";

// Mock data for a parallel workflow example
const mockParallelWorkflow: WorkflowDetail = {
  id: "4",
  name: "Parallel Data Processing",
  description: "A workflow with parallel processing paths",
  workflowType: "parallel",
  createdBy: "user-1",
  organizationId: "org-1",
  status: "active",
  config: {},
  workflowMetadata: {},
  version: "1.0.0",
  tags: ["parallel", "processing"],
  aiEnabled: false,
  createdAt: "2025-01-20T10:00:00",
  updatedAt: "2025-01-20T10:00:00",
  lastExecutedAt: "2025-01-20T10:00:00",
  steps: [
    { 
      id: "s1", 
      name: "Start", 
      description: "Initial step", 
      type: "start",
      workflowId: "4",
      stepOrder: 1,
      status: "completed",
      isRequired: true
    },
    { 
      id: "s2", 
      name: "Collect Data", 
      description: "Fork: Collect data from various sources", 
      type: "process",
      workflowId: "4",
      stepOrder: 2,
      status: "completed",
      isRequired: true
    },
    { 
      id: "s3a", 
      name: "Process Images", 
      description: "Process image data", 
      type: "process",
      workflowId: "4",
      stepOrder: 3,
      status: "active",
      isRequired: true
    },
    { 
      id: "s3b", 
      name: "Process Text", 
      description: "Process text data", 
      type: "process",
      workflowId: "4",
      stepOrder: 3,
      status: "active",
      isRequired: true
    },
    { 
      id: "s3c", 
      name: "Process Metrics", 
      description: "Process numerical metrics", 
      type: "process",
      workflowId: "4",
      stepOrder: 3,
      status: "pending",
      isRequired: true
    },
    { 
      id: "s4a", 
      name: "Analyze Images", 
      description: "Analyze processed images", 
      type: "process",
      workflowId: "4",
      stepOrder: 4,
      status: "pending",
      isRequired: true
    },
    { 
      id: "s4b", 
      name: "Analyze Text", 
      description: "Analyze processed text", 
      type: "process",
      workflowId: "4",
      stepOrder: 4,
      status: "pending",
      isRequired: true
    },
    { 
      id: "s5", 
      name: "Merge Results", 
      description: "Join: Combine all analysis results", 
      type: "decision",
      workflowId: "4",
      stepOrder: 5,
      status: "pending",
      isRequired: true
    },
    { 
      id: "s6", 
      name: "Generate Report", 
      description: "Create final report", 
      type: "process",
      workflowId: "4",
      stepOrder: 6,
      status: "pending",
      isRequired: true
    },
    { 
      id: "s7", 
      name: "Complete", 
      description: "Workflow completed", 
      type: "end",
      workflowId: "4",
      stepOrder: 7,
      status: "pending",
      isRequired: true
    }
  ],
  transitions: [
    // Start to fork
    { id: "t1", from: "s1", to: "s2", workflowId: "4" },
    
    // Fork to parallel paths
    { id: "t2a", from: "s2", to: "s3a", workflowId: "4" },
    { id: "t2b", from: "s2", to: "s3b", workflowId: "4" },
    { id: "t2c", from: "s2", to: "s3c", workflowId: "4" },
    
    // Parallel paths continue
    { id: "t3a", from: "s3a", to: "s4a", workflowId: "4" },
    { id: "t3b", from: "s3b", to: "s4b", workflowId: "4" },
    
    // Join paths
    { id: "t4a", from: "s4a", to: "s5", workflowId: "4" },
    { id: "t4b", from: "s4b", to: "s5", workflowId: "4" },
    { id: "t4c", from: "s3c", to: "s5", workflowId: "4" },
    
    // Continue to end
    { id: "t5", from: "s5", to: "s6", workflowId: "4" },
    { id: "t6", from: "s6", to: "s7", workflowId: "4" }
  ]
};

// Define types for the workflow layout
interface WorkflowNode {
  id: string;
  parentIds: string[];
  childIds: string[];
  step: WorkflowStep;
  track: number;
  level: number;
  isCompleted: boolean;
}

interface WorkflowConnection {
  from: string;
  to: string;
  percent: number;
}

interface WorkflowLayoutResult {
  grid: (WorkflowNode | null)[][];
  workflowGraph: Record<string, WorkflowNode>;
  horizontalLines: number[][];
  verticalLines: number[][];
  connectionLines: WorkflowConnection[];
  tracks: number;
  getNodePosition: (stepId: string) => { level: number; track: number } | null;
  isForkNode: (stepId: string) => boolean;
  isJoinNode: (stepId: string) => boolean;
}

interface ParallelWorkflowDetailProps {
  darkMode?: boolean;
  mockData?: WorkflowDetail; // For testing
}

export default function ParallelWorkflowDetailPage({ 
  darkMode = false,
  mockData
}: ParallelWorkflowDetailProps) {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [workflow, setWorkflow] = useState<WorkflowDetail | null>(null);
  const [activeStep, setActiveStep] = useState<string | null>(null);
  const [completedSteps, setCompletedSteps] = useState<string[]>([]);
  const [expandedStep, setExpandedStep] = useState<string | null>(null);
  const [lastCompletedStep, setLastCompletedStep] = useState<string | null>(null);
  const [workflowLayout, setWorkflowLayout] = useState<WorkflowLayoutResult | null>(null);

  useEffect(() => {
    // Use mock data provided as prop or the default mock parallel workflow
    const workflowData = mockData || mockParallelWorkflow;
    setWorkflow(workflowData);
    
    // Set initial completed steps
    const initialCompletedSteps = workflowData.steps
      .filter(step => step.status === "completed")
      .map(step => step.id);
    
    setCompletedSteps(initialCompletedSteps);
    
    // Set active step
    const activeSteps = workflowData.steps.filter(step => step.status === "active");
    if (activeSteps.length > 0) {
      setActiveStep(activeSteps[0].id);
    } else if (workflowData.steps.length > 0) {
      setActiveStep(workflowData.steps[0].id);
    }
    
    // Set last completed step
    if (initialCompletedSteps.length > 0) {
      setLastCompletedStep(initialCompletedSteps[initialCompletedSteps.length - 1]);
    }
  }, [id, mockData]);

  useEffect(() => {
    if (workflow) {
      // Calculate the workflow layout using our parallel workflow animation utility
      const layout = calculateParallelAnimations(
        workflow.steps,
        workflow.transitions,
        completedSteps
      );
      setWorkflowLayout(layout);
    }
  }, [workflow, completedSteps]);

  const handleBack = () => {
    navigate(-1); // Go back to the previous page
  };

  const getStepIcon = (type: string, stepId: string, isCompleted: boolean = false) => {
    if (isCompleted) {
      return <CheckCircle2 className="h-6 w-6 text-white" />;
    }
    
    // Check if this is a fork or join node
    if (workflowLayout) {
      if (workflowLayout.isForkNode(stepId)) {
        return <GitBranch className="h-6 w-6 text-indigo-500" />;
      }
      
      if (workflowLayout.isJoinNode(stepId)) {
        return <GitMerge className="h-6 w-6 text-indigo-500" />;
      }
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
      default:
        return <CircleDot className="h-6 w-6" />;
    }
  };

  const getStepBadgeStyles = (type: string, stepId: string) => {
    // Special styling for fork/join nodes
    if (workflowLayout) {
      if (workflowLayout.isForkNode(stepId)) {
        return "text-indigo-600 border-indigo-200 bg-indigo-50 dark:text-indigo-400 dark:border-indigo-950 dark:bg-indigo-950 dark:bg-opacity-20";
      }
      
      if (workflowLayout.isJoinNode(stepId)) {
        return "text-indigo-600 border-indigo-200 bg-indigo-50 dark:text-indigo-400 dark:border-indigo-950 dark:bg-indigo-950 dark:bg-opacity-20";
      }
    }
    
    switch (type) {
      case "start":
        return "text-emerald-600 border-emerald-200 bg-emerald-50 dark:text-emerald-400 dark:border-emerald-950 dark:bg-emerald-950 dark:bg-opacity-20";
      case "process":
        return "text-primary border-primary/20 bg-primary/10 dark:text-primary dark:border-primary/20 dark:bg-primary/10";
      case "decision":
        return "text-amber-600 border-amber-200 bg-amber-50 dark:text-amber-400 dark:border-amber-950 dark:bg-amber-950 dark:bg-opacity-20";
      case "end":
        return "text-destructive border-destructive/20 bg-destructive/10 dark:text-destructive dark:border-destructive/20 dark:bg-destructive/10";
      default:
        return "text-muted-foreground border-border bg-muted dark:text-muted-foreground dark:border-border dark:bg-muted";
    }
  };

  const getStepRingStyle = (type: string, stepId: string) => {
    // Special styling for fork/join nodes
    if (workflowLayout) {
      if (workflowLayout.isForkNode(stepId)) {
        return "ring-indigo-500";
      }
      
      if (workflowLayout.isJoinNode(stepId)) {
        return "ring-indigo-500";
      }
    }
    
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

  const handleToggleStepCompletion = (stepId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    const newCompletedSteps = completedSteps.includes(stepId)
      ? completedSteps.filter(id => id !== stepId)
      : [...completedSteps, stepId];
      
    setCompletedSteps(newCompletedSteps);
    
    if (!completedSteps.includes(stepId)) {
      setLastCompletedStep(stepId);
    } else if (stepId === lastCompletedStep) {
      // If we're un-completing the last step, find the new "last" completed step
      const remainingCompleted = newCompletedSteps.filter(id => id !== stepId);
      setLastCompletedStep(remainingCompleted.length > 0 ? remainingCompleted[remainingCompleted.length - 1] : null);
    }
  };

  const handleExpandStep = (stepId: string) => {
    setExpandedStep(expandedStep === stepId ? null : stepId);
    setActiveStep(stepId);
  };

  if (!workflow || !workflowLayout) {
    return (
      <div className="flex flex-1 flex-col p-6 h-[calc(100vh-32px)] overflow-hidden">
        <p>Loading workflow...</p>
      </div>
    );
  }

  // Calculate progress percentage
  const progressPercentage = workflow.steps.length > 0
    ? (completedSteps.length / workflow.steps.length) * 100
    : 0;

  return (
    <div className="flex flex-1 flex-col gap-4 p-6 h-[calc(100vh-32px)] overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between">
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

        <div className="flex items-center gap-2">
          <Button 
            variant="outline" 
            size="sm" 
            className="gap-1.5 h-8"
          >
            <CheckIcon className="h-3.5 w-3.5" />
            Auto-Run
          </Button>
        </div>
      </div>

      {/* Main content container */}
      <div className="flex-1 flex flex-col min-h-0">
        {/* Workflow Status Bar */}
        <div className="flex items-center justify-between mb-3 px-1 py-1.5">
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
          
          <div className="flex items-center gap-2">
            <span className="text-xs font-medium">{Math.round(progressPercentage)}%</span>
            <div className="h-1.5 w-24 bg-muted rounded-full overflow-hidden">
              <motion.div 
                className={cn(
                  "h-full rounded-full",
                  completedSteps.length === workflow.steps.length
                    ? "bg-emerald-500 dark:bg-emerald-500"
                    : "bg-primary"
                )}
                initial={{ width: 0 }}
                animate={{ width: `${progressPercentage}%` }}
                transition={{ duration: 0.5, ease: "easeOut" }}
              />
            </div>
            <span className="text-xs text-muted-foreground">
              {completedSteps.length}/{workflow.steps.length}
            </span>
          </div>
        </div>

        {/* Workflow Steps Container */}
        <div className="flex-1 bg-background border border-border rounded-xl overflow-hidden flex flex-col min-h-0">
          <div className="p-4 border-b border-border flex items-center justify-between bg-card">
            <h3 className="font-medium">Parallel Workflow</h3>
            <div className="flex items-center text-sm">
              <span className="mr-2 text-muted-foreground">Completion Flow</span>
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
            {/* Parallel Workflow Grid */}
            <div className="relative mt-8 mb-4 px-4">
              {/* Render the workflow grid */}
              {workflowLayout.grid.map((row: (WorkflowNode | null)[], levelIndex: number) => (
                <div key={`level-${levelIndex}`} className="relative mb-24">
                  {/* Level label */}
                  <div className="absolute -left-12 top-1/2 transform -translate-y-1/2 text-xs font-medium text-muted-foreground">
                    Level {levelIndex}
                  </div>
                  
                  {/* Nodes at this level */}
                  <div className="flex justify-around items-center relative">
                    {/* Horizontal connection line */}
                    {levelIndex > 0 && (
                      <div 
                        className="absolute left-0 right-0 h-1.5 bg-border dark:bg-muted rounded-full z-0"
                        style={{ top: '50%', transform: 'translateY(-50%)' }}
                      />
                    )}
                    
                    {/* Render nodes */}
                    {row.map((node: WorkflowNode | null, trackIndex: number) => {
                      if (!node) return (
                        <div 
                          key={`empty-${levelIndex}-${trackIndex}`} 
                          className="w-48 py-8 opacity-0" // Invisible placeholder
                        />
                      );
                      
                      const step = node.step;
                      const isCompleted = completedSteps.includes(step.id);
                      const isExpanded = expandedStep === step.id;
                      const isForkNode = workflowLayout.isForkNode(step.id);
                      const isJoinNode = workflowLayout.isJoinNode(step.id);
                      
                      return (
                        <motion.div 
                          key={step.id} 
                          className="flex flex-col items-center relative z-10 w-48 mx-2"
                          initial={{ opacity: 0, y: 20 }}
                          animate={{ opacity: 1, y: 0 }}
                          transition={{ 
                            duration: 0.3, 
                            delay: (levelIndex * 0.2) + (trackIndex * 0.1) 
                          }}
                        >
                          {/* Step Node */}
                          <motion.div 
                            className={cn(
                              "relative w-full p-5 rounded-xl shadow-md transition-all duration-200",
                              "bg-card border border-border",
                              activeStep === step.id && `ring-2 ${getStepRingStyle(step.type, step.id)} ring-opacity-70`,
                              isCompleted && "border-primary/30 dark:border-primary/30",
                              "cursor-pointer group z-10"
                            )}
                            layout
                            onClick={() => handleExpandStep(step.id)}
                            whileHover={{ scale: 1.02, y: -2 }}
                            transition={{ 
                              layout: { duration: 0.2 },
                              scale: { duration: 0.1 }
                            }}
                          >
                            {/* Step Icon */}
                            <motion.div
                              className={cn(
                                "absolute top-0 rounded-full p-2.5 shadow-md border border-border",
                                "bg-background group-hover:scale-110 transition-transform",
                                isCompleted ? "bg-emerald-600" : "",
                                isForkNode && !isCompleted ? "bg-indigo-100 dark:bg-indigo-900/30" : "",
                                isJoinNode && !isCompleted ? "bg-indigo-100 dark:bg-indigo-900/30" : ""
                              )}
                              onClick={(e) => handleToggleStepCompletion(step.id, e)}
                              style={{ 
                                left: '50%',
                                transform: 'translate(-50%, -50%)',
                                zIndex: 30
                              }}
                            >
                              {getStepIcon(step.type, step.id, isCompleted)}
                            </motion.div>
                            
                            <div className="w-full pt-4">
                              <div className="flex justify-between items-start">
                                <h3 className={cn(
                                  "font-medium text-base text-card-foreground",
                                  isCompleted && "text-primary dark:text-primary"
                                )}>{step.name}</h3>
                                <Badge variant="outline" className={cn(
                                  "text-xs capitalize",
                                  getStepBadgeStyles(step.type, step.id)
                                )}>
                                  {isForkNode ? "Fork" : isJoinNode ? "Join" : step.type}
                                </Badge>
                              </div>
                              <p className="text-sm text-muted-foreground mt-1.5">{step.description}</p>
                              
                              {/* Expanded Actions */}
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
                                        >
                                          <CheckIcon className="h-3.5 w-3.5 mr-1.5" />
                                          {isCompleted ? "Mark Incomplete" : "Mark Complete"}
                                        </Button>
                                      </TooltipTrigger>
                                      <TooltipContent>
                                        <p>Toggle step completion status</p>
                                      </TooltipContent>
                                    </Tooltip>
                                  </TooltipProvider>
                                  
                                  <TooltipProvider>
                                    <Tooltip>
                                      <TooltipTrigger asChild>
                                        <Button 
                                          size="sm" 
                                          variant="ghost"
                                          className="h-8 w-8 p-0 text-muted-foreground"
                                          onClick={(e) => e.stopPropagation()}
                                        >
                                          <MoreHorizontal className="h-4 w-4" />
                                        </Button>
                                      </TooltipTrigger>
                                      <TooltipContent>
                                        <p>More actions</p>
                                      </TooltipContent>
                                    </Tooltip>
                                  </TooltipProvider>
                                </motion.div>
                              )}
                            </div>
                          </motion.div>
                          
                          {/* Render connection lines for this node */}
                          {workflowLayout.connectionLines
                            .filter(conn => conn.from === step.id)
                            .map((conn: WorkflowConnection) => {
                              const toNode = workflowLayout.workflowGraph[conn.to];
                              if (!toNode) return null;
                              
                              // Determine if this is a vertical connection
                              const isVertical = toNode.level > node.level;
                              
                              // For vertical connections
                              if (isVertical) {
                                return (
                                  <div 
                                    key={`conn-${conn.from}-${conn.to}`}
                                    className="absolute w-1.5 bg-border dark:bg-muted rounded-full overflow-hidden"
                                    style={{
                                      left: '50%',
                                      top: '100%',
                                      height: '80px',
                                      transform: 'translateX(-50%)',
                                      zIndex: 5
                                    }}
                                  >
                                    <motion.div 
                                      className="w-full bg-primary rounded-full"
                                      style={{ 
                                        height: `${conn.percent}%`,
                                        transformOrigin: 'top'
                                      }}
                                      initial={{ height: 0 }}
                                      animate={{ height: `${conn.percent}%` }}
                                      transition={{ duration: 0.5, ease: "easeOut" }}
                                    />
                                  </div>
                                );
                              }
                              
                              // For horizontal connections (same level, between tracks)
                              return null; // These are handled by the horizontal line
                            })}
                        </motion.div>
                      );
                    })}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
} 