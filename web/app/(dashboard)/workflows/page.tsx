"use client";

import React, { useCallback } from "react";
import { WorkflowBuilder } from "@/components/recovery/builder/WorkflowBuilder";

export default function WorkflowsPage() {
  const handleSave = useCallback((steps: any) => {
    console.log("Saved steps:", steps);
  }, []);

  return (
    <div className="flex flex-col gap-6 p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">
            Workflow Builder
          </h1>
          <p className="text-muted-foreground mt-1">
            Design dynamic recovery strategies using the visual workflow editor.
          </p>
        </div>
      </div>

      <WorkflowBuilder onSave={handleSave} />
    </div>
  );
}
