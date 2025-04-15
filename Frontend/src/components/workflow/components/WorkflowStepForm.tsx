import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardFooter } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { StepType, WorkflowStepRequest } from "@/components/workflow/types";
import { X } from "lucide-react";

interface WorkflowStepFormProps {
  onClose: () => void;
  onSubmit: (data: Partial<WorkflowStepRequest>) => void;
  isLoading?: boolean;
}

const availableStepTypes: StepType[] = [
  "manual",
  "automated",
  "approval",
  "notification",
  "integration",
  "decision",
  "ai_task",
];

export default function WorkflowStepForm({ onClose, onSubmit, isLoading }: WorkflowStepFormProps) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [type, setType] = useState<StepType>("manual");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit({
      name,
      description,
      step_type: type,
    });
  };

  return (
    <div className="fixed inset-0 bg-black/60 z-50 flex items-center justify-center">
      <Card className="w-[90vw] max-w-lg relative">
        <button onClick={onClose} className="absolute top-4 right-4 text-muted-foreground hover:text-foreground">
          <X className="h-5 w-5" />
        </button>
        <CardHeader>
          <CardTitle>Add New Step</CardTitle>
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
              <Label htmlFor="type">Step Type</Label>
              <Select onValueChange={(value) => setType(value as StepType)} defaultValue={type}>
                <SelectTrigger id="type">
                  <SelectValue placeholder="Select a step type" />
                </SelectTrigger>
                <SelectContent>
                  {availableStepTypes.map((stepType) => (
                    <SelectItem key={stepType} value={stepType} className="capitalize">
                      {stepType.replace("_", " ")}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </CardContent>
          <CardFooter className="flex justify-end gap-2">
            <Button type="button" variant="ghost" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? "Adding..." : "Add Step"}
            </Button>
          </CardFooter>
        </form>
      </Card>
    </div>
  );
} 