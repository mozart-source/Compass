import React, { useState } from 'react';
import { X } from 'lucide-react';
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { WorkflowType } from '@/components/workflow/types';
import { useCreateWorkflow } from '@/components/workflow/hooks';

interface WorkflowFormProps {
  onClose: () => void;
  onSubmit: (data: WorkflowFormData) => void;
  initialData?: Partial<WorkflowFormData>;
  darkMode?: boolean;
}

interface WorkflowFormData {
  name: string;
  description: string;
  workflow_type: WorkflowType;
  ai_enabled: boolean;
  tags: string[];
  estimatedDuration?: number;
  deadline?: string;
}

const WorkflowForm: React.FC<WorkflowFormProps> = ({
  onClose,
  onSubmit,
  initialData,
  darkMode = false
}) => {
  const [isClosing, setIsClosing] = useState(false);
  const [name, setName] = useState(initialData?.name || '');
  const [description, setDescription] = useState(initialData?.description || '');
  const [workflowType, setWorkflowType] = useState<WorkflowType>(initialData?.workflow_type || 'sequential');
  const [aiEnabled, setAiEnabled] = useState(initialData?.ai_enabled || false);

  const createWorkflow = useCreateWorkflow();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    const formData: WorkflowFormData = {
      name,
      description,
      workflow_type: workflowType,
      ai_enabled: aiEnabled,
      tags: [],
    };

    onSubmit(formData);
  };

  const handleClose = () => {
    setIsClosing(true);
    setTimeout(onClose, 300);
  };

  return (
    <div className={cn(
      "fixed inset-0 z-50 flex items-center justify-center bg-black/50",
      "animate-fade-in",
      isClosing && "animate-fade-out"
    )}>
      <div className={cn(
        "w-full max-w-md rounded-lg bg-card p-6 relative",
        "animate-fade-in",
        isClosing && "animate-fade-out"
      )}>
        <button
          onClick={handleClose}
          className="absolute right-4 top-4 text-muted-foreground hover:text-foreground"
        >
          <X className="w-5 h-5" />
        </button>

        <h2 className="text-xl font-semibold mb-6">Create New Workflow</h2>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Workflow name"
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Workflow description"
              className="min-h-[100px]"
            />
          </div>

          <div className="space-y-2">
            <Label>Workflow Type</Label>
            <Select
              value={workflowType}
              onValueChange={(value: WorkflowType) => setWorkflowType(value)}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select workflow type" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="sequential">Sequential</SelectItem>
                <SelectItem value="parallel">Parallel</SelectItem>
                <SelectItem value="conditional">Conditional</SelectItem>
                <SelectItem value="ai_driven">AI Driven</SelectItem>
                <SelectItem value="hybrid">Hybrid</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex justify-end gap-3 mt-6">
            <Button type="button" variant="outline" onClick={handleClose} className="px-4 py-2">
              Cancel
            </Button>
            <Button type="submit" className="px-4 py-2" disabled={createWorkflow.isPending}>
              {createWorkflow.isPending ? 'Creating...' : 'Create Workflow'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default WorkflowForm; 