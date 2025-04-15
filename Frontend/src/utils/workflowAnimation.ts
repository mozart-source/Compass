import { WorkflowStep } from "@/components/workflow/types";

export interface StepConnectionInfo {
  rowIndex: number;
  colIndex: number;
  stepId: string;
  isCompleted: boolean;
}

export function calculateLineAnimations(
  groupedSteps: WorkflowStep[][],
  completedSteps: string[]
) {
  // Create a map of step connections
  const stepMap = new Map<string, StepConnectionInfo>();
  
  // Populate the map
  groupedSteps.forEach((row, rowIndex) => {
    row.forEach((step, colIndex) => {
      stepMap.set(step.id, {
        rowIndex,
        colIndex,
        stepId: step.id,
        isCompleted: completedSteps.includes(step.id)
      });
    });
  });
  
  // Find the last completed step info across the entire workflow
  const lastCompletedStepInfo = findLastCompletedStepInfo(groupedSteps, completedSteps);
  
  // Calculate the horizontal line animations with zigzag connection awareness
  const horizontalLines = groupedSteps.map((row, rowIndex) => {
    // If we haven't reached this row yet in the completion flow
    if (lastCompletedStepInfo && rowIndex > lastCompletedStepInfo.rowIndex) {
      return 0;
    }
    
    // If this row is before our last completed step's row, it's fully completed
    if (lastCompletedStepInfo && rowIndex < lastCompletedStepInfo.rowIndex) {
      return 100;
    }
    
    // For the row containing the last completed step
    const completedInRow = row.filter(step => completedSteps.includes(step.id));
    
    if (completedInRow.length === 0) return 0;
    if (completedInRow.length === row.length) return 100;
    
    const isRtl = rowIndex % 2 !== 0;
    // For RTL rows, we need to map the visual positions differently
    // In the UI for RTL rows: 
    // - visual left = node at index 2
    // - visual right = node at index 0
    
    if (isRtl) {
      // For RTL rows, the nodes visually go from right to left (index 0 is rightmost)
      // Calculate how many "positions" from the left edge are completed
      const visualPositions = row.map((step, idx) => ({
        step,
        // Invert the indices for RTL rows (0 becomes 2, 1 remains 1, 2 becomes 0)
        visualIndex: row.length - 1 - idx,
        completed: completedSteps.includes(step.id)
      }));
      
      // Find the rightmost completed node (lowest visualIndex)
      const rightmostCompletedVisualIndex = Math.min(
        ...visualPositions
          .filter(p => p.completed)
          .map(p => p.visualIndex)
      );
      
      // Calculate percentage - for RTL rows, progress starts from right (0%) to left (100%)
      return ((row.length - 1 - rightmostCompletedVisualIndex) / (row.length - 1)) * 100;
    } else {
      // For LTR rows, progress goes from left to right (0% to 100%)
      // Get indices of steps that are completed
      const completedIndices = row
        .map((step, idx) => ({ step, idx }))
        .filter(({ step }) => completedSteps.includes(step.id))
        .map(({ idx }) => idx);
      
      const lastCompletedIndex = Math.max(...completedIndices);
      return (lastCompletedIndex / (row.length - 1)) * 100;
    }
  });
  
  // Enhanced vertical line calculation that integrates with horizontal lines
  const verticalLines = groupedSteps.slice(1).map((_, rowIndex) => {
    // rowIndex here is the destination row (1-based since we sliced)
    const actualRowIndex = rowIndex + 1;
    
    // Get the completion status of the current row's horizontal line
    const currentRowCompletion = horizontalLines[actualRowIndex];
    
    // Get the completion status of the previous row's horizontal line
    const prevRowCompletion = horizontalLines[actualRowIndex - 1];
    
    // Get current and previous rows
    const prevRow = groupedSteps[actualRowIndex - 1];
    const currentRow = groupedSteps[actualRowIndex];
    
    // Determine zigzag connection points
    const isPrevRowLTR = (actualRowIndex - 1) % 2 === 0;
    
    // For LTR rows, the connection is from the last node (rightmost, index: length-1)
    // For RTL rows, the connection is from the first visual node (leftmost, but still index 0 in the array)
    const prevRowConnectionIdx = isPrevRowLTR ? prevRow.length - 1 : 0;
    
    // The visual position of connection points between rows
    const prevRowConnectionStep = prevRow[prevRowConnectionIdx];
    const currentRowConnectionStep = currentRow[0]; // Always the first node in array for any row
    
    // Check if either the first node of current row OR a node in the current row is completed
    const isAnyCurrentRowStepCompleted = currentRow.some(step => completedSteps.includes(step.id));
    
    // Special case: When first and last nodes of workflow are completed
    // This ensures the entire workflow path shows connected
    if (rowIndex === 0 && completedSteps.length > 0) {
      // Get first node of the entire workflow
      const firstNode = groupedSteps[0][0];
      
      // Find last row and its last node
      const lastRowIndex = groupedSteps.length - 1;
      const lastRow = groupedSteps[lastRowIndex];
      const lastNode = lastRow[lastRow.length - 1];
      
      // If both first and last nodes are completed, ensure all vertical connections show
      if (completedSteps.includes(firstNode.id) && completedSteps.includes(lastNode.id)) {
        return 100;
      }
    }
    
    // If the previous row isn't started, vertical line is also 0
    if (prevRowCompletion === 0 && !isAnyCurrentRowStepCompleted) {
      return 0; 
    }
    
    // If previous row is completed (100%)
    if (prevRowCompletion === 100) {
      // If any step in the current row is completed, show full vertical line
      if (isAnyCurrentRowStepCompleted) {
        return 100; // Full vertical line
      } else if (completedSteps.includes(prevRowConnectionStep.id) && completedSteps.includes(currentRowConnectionStep.id)) {
        // Only show if both connection points are completed
        return 100; // Full vertical line
      }
    }
    
    // If previous row is partially completed and the connection point is completed
    if (prevRowCompletion > 0 && completedSteps.includes(prevRowConnectionStep.id)) {
      // If any step in the current row is completed, show full vertical line
      if (isAnyCurrentRowStepCompleted && completedSteps.includes(currentRowConnectionStep.id)) {
        return 100; // Complete the vertical connection
      } else if (completedSteps.includes(currentRowConnectionStep.id)) {
        // Only show halfway if the current row's first node is also completed
        return 50; // Halfway vertical line
      }
      return 0; // Don't show line unless there's a valid path
    }
    
    // If current row has any completed steps but previous connection isn't completed
    if (isAnyCurrentRowStepCompleted && completedSteps.includes(currentRowConnectionStep.id)) {
      return 50; // Show halfway vertical line starting from the bottom
    }
    
    return 0;
  });
  
  // Get the connection status for the most recently completed step
  const getConnectionPath = (stepId: string): number[][] => {
    if (!stepMap.has(stepId)) return [];
    
    const { rowIndex, colIndex } = stepMap.get(stepId)!;
    const path: number[][] = [];
    
    // Add all rows up to the current one
    for (let r = 0; r <= rowIndex; r++) {
      // For each row, determine how far the highlight should go
      const rowSteps = groupedSteps[r];
      const isRtl = r % 2 !== 0;
      
      // If this is the target row, highlight up to the target column
      if (r === rowIndex) {
        // For RTL rows, we need to calculate the visual position differently
        // In RTL rows, colIndex 0 is visually on the right, and max index is on the left
        const visualColIndex = isRtl ? rowSteps.length - 1 - colIndex : colIndex;
        path.push([r, visualColIndex]);
      }
      // Otherwise highlight the full row if it should be highlighted
      else if (rowSteps.every(step => completedSteps.includes(step.id))) {
        // For completed rows, always push the rightmost visual node for RTL rows
        // and the leftmost for LTR rows
        const maxVisualIndex = isRtl ? 0 : rowSteps.length - 1;
        path.push([r, maxVisualIndex]);
      }
      // Or partially highlight the row based on completed steps
      else {
        const completedInRow = rowSteps.filter(step => completedSteps.includes(step.id));
        if (completedInRow.length > 0) {
          if (isRtl) {
            // For RTL rows, find the leftmost completed node (highest array index)
            const leftmostCompletedIndex = Math.max(
              ...rowSteps
                .map((step, idx) => ({ step, idx }))
                .filter(({ step }) => completedSteps.includes(step.id))
                .map(({ idx }) => idx)
            );
            // Convert to visual position
            const visualIndex = rowSteps.length - 1 - leftmostCompletedIndex;
            path.push([r, visualIndex]);
          } else {
            // For LTR rows, find the rightmost completed node (highest array index)
            const rightmostCompletedIndex = Math.max(
              ...rowSteps
                .map((step, idx) => ({ step, idx }))
                .filter(({ step }) => completedSteps.includes(step.id))
                .map(({ idx }) => idx)
            );
            path.push([r, rightmostCompletedIndex]);
          }
        }
      }
    }
    
    return path;
  };
  
  return {
    horizontalLines,
    verticalLines,
    getConnectionPath
  };
}

// Helper function to find information about the last completed step
function findLastCompletedStepInfo(
  groupedSteps: WorkflowStep[][],
  completedSteps: string[]
): { rowIndex: number; colIndex: number; stepId: string } | null {
  if (completedSteps.length === 0) return null;
  
  const completedStepIds = new Set(completedSteps);
  
  let lastCompletedStep: WorkflowStep | undefined;
  let lastRowIndex = -1;
  let lastColIndex = -1;
  
  // Find the last completed step in the workflow order
  groupedSteps.forEach((row, rowIndex) => {
    row.forEach((step, colIndex) => {
      if (completedStepIds.has(step.id)) {
        lastCompletedStep = step;
        lastRowIndex = rowIndex;
        lastColIndex = colIndex;
      }
    });
  });
  
  if (!lastCompletedStep) return null;
  
  return {
    rowIndex: lastRowIndex,
    colIndex: lastColIndex,
    stepId: lastCompletedStep.id
  };
} 