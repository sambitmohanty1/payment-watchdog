"use client";

import React, { useState } from "react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import { MaterialIcon } from "@/components/ui/MaterialIcon";

export type StepType =
  | "retry_payment"
  | "send_email"
  | "send_sms"
  | "wait"
  | "conditional";

export interface WorkflowStep {
  id: string;
  type: StepType;
  name: string;
  config?: any;
}

interface WorkflowBuilderProps {
  initialSteps?: WorkflowStep[];
  onSave?: (steps: WorkflowStep[]) => void;
}

const STEP_COLORS: Record<StepType, string> = {
  retry_payment:
    "bg-blue-100 border-blue-300 dark:bg-blue-900 dark:border-blue-700",
  send_email:
    "bg-green-100 border-green-300 dark:bg-green-900 dark:border-green-700",
  send_sms:
    "bg-purple-100 border-purple-300 dark:bg-purple-900 dark:border-purple-700",
  wait: "bg-orange-100 border-orange-300 dark:bg-orange-900 dark:border-orange-700",
  conditional:
    "bg-yellow-100 border-yellow-300 dark:bg-yellow-900 dark:border-yellow-700",
};

const STEP_LABELS: Record<StepType, string> = {
  retry_payment: "Retry Payment",
  send_email: "Send Email",
  send_sms: "Send SMS",
  wait: "Wait",
  conditional: "Condition",
};

export function WorkflowBuilder({
  initialSteps = [],
  onSave,
}: WorkflowBuilderProps) {
  const [steps, setSteps] = useState<WorkflowStep[]>(
    initialSteps.length
      ? initialSteps
      : [
          {
            id: "1",
            type: "wait",
            name: "Wait 3 Days",
            config: { delayMinutes: 4320 },
          },
          { id: "2", type: "retry_payment", name: "First Attempt", config: {} },
          {
            id: "3",
            type: "send_email",
            name: "Failure Notification",
            config: { templateId: "default" },
          },
        ],
  );
  const [draggedStepIdx, setDraggedStepIdx] = useState<number | null>(null);

  const addStep = (type: StepType) => {
    const newStep: WorkflowStep = {
      id: Math.random().toString(36).substring(7),
      type,
      name: `New ${STEP_LABELS[type]}`,
    };
    setSteps([...steps, newStep]);
  };

  const removeStep = (id: string) => {
    setSteps(steps.filter((s) => s.id !== id));
  };

  const moveStep = (fromIdx: number, toIdx: number) => {
    const newSteps = [...steps];
    const [moved] = newSteps.splice(fromIdx, 1);
    newSteps.splice(toIdx, 0, moved);
    setSteps(newSteps);
  };

  const handleDragStart = (e: React.DragEvent, idx: number) => {
    setDraggedStepIdx(idx);
    e.dataTransfer.effectAllowed = "move";
  };

  const handleDragOver = (e: React.DragEvent, idx: number) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = "move";
  };

  const handleDrop = (e: React.DragEvent, toIdx: number) => {
    e.preventDefault();
    if (draggedStepIdx !== null && draggedStepIdx !== toIdx) {
      moveStep(draggedStepIdx, toIdx);
    }
    setDraggedStepIdx(null);
  };

  const handleSave = () => {
    if (onSave) {
      onSave(steps);
    }
  };

  return (
    <div className="flex flex-col md:flex-row gap-6">
      {/* Sidebar / Tools */}
      <Card className="w-full md:w-64 h-fit">
        <CardHeader>
          <CardTitle className="text-lg">Add Step</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-2">
          {Object.entries(STEP_LABELS).map(([type, label]) => (
            <Button
              key={type}
              variant="outline"
              className="justify-start w-full"
              onClick={() => addStep(type as StepType)}
            >
              <MaterialIcon name="add" className="mr-2 h-4 w-4" />
              {label}
            </Button>
          ))}
        </CardContent>
      </Card>

      {/* Canvas */}
      <Card className="flex-1">
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>Workflow Execution Path</CardTitle>
          <Button onClick={handleSave}>Save Workflow</Button>
        </CardHeader>
        <CardContent>
          {steps.length === 0 ? (
            <div className="flex flex-col items-center justify-center p-12 text-muted-foreground border-2 border-dashed rounded-lg">
              <MaterialIcon name="warning" className="h-12 w-12 mb-4 text-muted-foreground/50" />
              <p>
                No steps defined. Add steps from the sidebar to create a
                workflow.
              </p>
            </div>
          ) : (
            <div className="flex flex-col items-center w-full max-w-xl mx-auto py-8">
              {steps.map((step, idx) => (
                <React.Fragment key={step.id}>
                  {/* Step Node */}
                  <div
                    draggable
                    onDragStart={(e) => handleDragStart(e, idx)}
                    onDragOver={(e) => handleDragOver(e, idx)}
                    onDrop={(e) => handleDrop(e, idx)}
                    onDragEnd={() => setDraggedStepIdx(null)}
                    className={`
                      w-full relative flex items-center p-4 border-2 rounded-lg cursor-grab active:cursor-grabbing
                      ${STEP_COLORS[step.type]}
                      ${draggedStepIdx === idx ? "opacity-50" : "opacity-100"}
                      transition-opacity
                    `}
                  >
                    <div className="mr-4 text-muted-foreground">
                      <MaterialIcon name="drag_indicator" className="h-5 w-5" />
                    </div>

                    <div className="flex-1">
                      <div className="flex items-center justify-between mb-1">
                        <span className="text-xs font-semibold uppercase tracking-wider opacity-75">
                          {STEP_LABELS[step.type]}
                        </span>
                        <span className="text-xs opacity-50">
                          Step {idx + 1}
                        </span>
                      </div>
                      <h4 className="font-medium text-lg leading-none">
                        {step.name}
                      </h4>
                    </div>

                    <div className="ml-4 flex gap-2">
                      <Button variant="ghost" size="icon" className="h-8 w-8">
                        <MaterialIcon name="settings" className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-destructive hover:text-destructive hover:bg-destructive/10"
                        onClick={() => removeStep(step.id)}
                      >
                        <MaterialIcon name="delete" className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>

                  {/* Connector */}
                  {idx < steps.length - 1 && (
                    <div className="h-10 flex items-center justify-center">
                      <div className="w-0.5 h-full bg-border relative">
                        <MaterialIcon name="south" className="h-4 w-4 absolute -bottom-2 -left-[7px] text-border bg-background" />
                      </div>
                    </div>
                  )}
                </React.Fragment>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
