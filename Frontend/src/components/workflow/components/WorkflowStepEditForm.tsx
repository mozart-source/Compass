import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardFooter,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import Checkbox from "@/components/ui/checkbox";
import { WorkflowStep, WorkflowStepRequest } from "@/components/workflow/types";
import { X } from "lucide-react";

interface WorkflowStepEditFormProps {
  step: WorkflowStep;
  onClose: () => void;
  onSubmit: (data: Partial<WorkflowStepRequest>) => void;
  isLoading?: boolean;
}

export default function WorkflowStepEditForm({
  step,
  onClose,
  onSubmit,
  isLoading,
}: WorkflowStepEditFormProps) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [assignedTo, setAssignedTo] = useState("");
  const [isRequired, setIsRequired] = useState(false);

  useEffect(() => {
    if (step) {
      setName(step.name);
      setDescription(step.description);
      setAssignedTo(step.assignedTo || "");
      setIsRequired(step.isRequired);
    }
  }, [step]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit({
      name,
      description,
      assignedTo: assignedTo || undefined,
      isRequired,
    });
  };

  return (
    <div className="fixed inset-0 bg-black/60 z-50 flex items-center justify-center">
      <Card className="w-[90vw] max-w-lg relative">
        <button
          onClick={onClose}
          className="absolute top-4 right-4 text-muted-foreground hover:text-foreground"
        >
          <X className="h-5 w-5" />
        </button>
        <CardHeader>
          <CardTitle>Edit Step: {step.name}</CardTitle>
        </CardHeader>
        <form onSubmit={handleSubmit}>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name">Step Name</Label>
              <Input
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g., 'Review Document'"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="A short description of what this step does."
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="assignedTo">Assigned To (User ID)</Label>
              <Input
                id="assignedTo"
                value={assignedTo}
                onChange={(e) => setAssignedTo(e.target.value)}
                placeholder="Enter user UUID (optional)"
              />
            </div>
            <div className="flex items-center space-x-2">
              <Checkbox
                id="isRequired"
                checked={isRequired}
                onCheckedChange={(checked) => setIsRequired(Boolean(checked))}
              />
              <Label htmlFor="isRequired">Is Required</Label>
            </div>
          </CardContent>
          <CardFooter className="flex justify-end gap-2">
            <Button type="button" variant="ghost" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? "Saving..." : "Save Changes"}
            </Button>
          </CardFooter>
        </form>
      </Card>
    </div>
  );
} 