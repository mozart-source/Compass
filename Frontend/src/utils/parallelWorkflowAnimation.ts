import { WorkflowStep } from "@/components/workflow/types";

export interface StepConnectionInfo {
  rowIndex: number;
  colIndex: number;
  stepId: string;
  isCompleted: boolean;
}

interface WorkflowNode {
  id: string;
  parentIds: string[];
  childIds: string[];
  step: WorkflowStep;
  track: number;
  level: number;
  isCompleted: boolean;
}

export function calculateParallelAnimations(
  steps: WorkflowStep[],
  transitions: { id: string; from: string; to: string; condition?: string; workflowId: string }[],
  completedSteps: string[]
) {
  // Step 1: Build the workflow graph
  const workflowGraph = buildWorkflowGraph(steps, transitions, completedSteps);
  
  // Step 2: Identify start and end nodes
  const startNodes = Object.values(workflowGraph).filter(node => node.parentIds.length === 0);
  const endNodes = Object.values(workflowGraph).filter(node => node.childIds.length === 0);
  
  // Step 3: Assign levels to nodes (depth from start)
  assignLevels(workflowGraph, startNodes);
  
  // Step 4: Assign tracks (parallel paths)
  const tracks = assignTracks(workflowGraph);
  
  // Step 5: Group nodes into a grid based on track and level
  const grid = createNodeGrid(workflowGraph);
  
  // Step 6: Calculate line animations
  const { horizontalLines, verticalLines, connectionLines } = calculateLineAnimations(
    grid, 
    workflowGraph,
    transitions,
    completedSteps
  );
  
  return {
    grid,
    workflowGraph,
    horizontalLines,
    verticalLines,
    connectionLines,
    tracks,
    getNodePosition: (stepId: string) => {
      const node = workflowGraph[stepId];
      if (!node) return null;
      return { level: node.level, track: node.track };
    },
    isForkNode: (stepId: string) => {
      const node = workflowGraph[stepId];
      return node && node.childIds.length > 1;
    },
    isJoinNode: (stepId: string) => {
      const node = workflowGraph[stepId];
      return node && node.parentIds.length > 1;
    }
  };
}

// Build a graph representation of the workflow
function buildWorkflowGraph(
  steps: WorkflowStep[],
  transitions: { id: string; from: string; to: string; condition?: string; workflowId: string }[],
  completedSteps: string[]
): Record<string, WorkflowNode> {
  const graph: Record<string, WorkflowNode> = {};
  
  // Initialize nodes
  steps.forEach(step => {
    graph[step.id] = {
      id: step.id,
      parentIds: [],
      childIds: [],
      step,
      track: 0,
      level: 0,
      isCompleted: completedSteps.includes(step.id)
    };
  });
  
  // Add edge connections
  transitions.forEach(transition => {
    // Add the destination as a child of the source
    if (graph[transition.from]) {
      graph[transition.from].childIds.push(transition.to);
    }
    
    // Add the source as a parent of the destination
    if (graph[transition.to]) {
      graph[transition.to].parentIds.push(transition.from);
    }
  });
  
  return graph;
}

// Assign levels to nodes (distance from start node)
function assignLevels(
  graph: Record<string, WorkflowNode>,
  startNodes: WorkflowNode[]
) {
  const visited = new Set<string>();
  
  function dfs(node: WorkflowNode, level: number) {
    if (visited.has(node.id)) {
      // Only update level if the new level is deeper
      if (level > node.level) {
        node.level = level;
      }
      return;
    }
    
    visited.add(node.id);
    node.level = level;
    
    // Process children
    for (const childId of node.childIds) {
      const childNode = graph[childId];
      if (childNode) {
        dfs(childNode, level + 1);
      }
    }
  }
  
  // Start DFS from each start node
  startNodes.forEach(node => dfs(node, 0));
}

// Assign tracks to nodes (parallel paths)
function assignTracks(graph: Record<string, WorkflowNode>): number {
  let currentTrack = 0;
  const assigned = new Set<string>();
  
  // Group nodes by level
  const nodesByLevel: Record<number, WorkflowNode[]> = {};
  
  Object.values(graph).forEach(node => {
    if (!nodesByLevel[node.level]) {
      nodesByLevel[node.level] = [];
    }
    nodesByLevel[node.level].push(node);
  });
  
  // Sort levels
  const levels = Object.keys(nodesByLevel).map(Number).sort((a, b) => a - b);
  
  // Process nodes level by level
  levels.forEach(level => {
    const nodesAtLevel = nodesByLevel[level];
    
    // Group nodes by their parent constellation
    const trackGroups: Record<string, WorkflowNode[]> = {};
    
    nodesAtLevel.forEach(node => {
      if (assigned.has(node.id)) return;
      
      // For start nodes or nodes with a single parent
      if (node.parentIds.length <= 1) {
        const parentId = node.parentIds[0] || "root";
        const parentNode = graph[parentId];
        const trackKey = parentId;
        
        if (!trackGroups[trackKey]) {
          trackGroups[trackKey] = [];
        }
        trackGroups[trackKey].push(node);
      }
      // For join nodes (multiple parents)
      else {
        // Create a unique key based on sorted parent IDs
        const trackKey = [...node.parentIds].sort().join("_");
        
        if (!trackGroups[trackKey]) {
          trackGroups[trackKey] = [];
        }
        trackGroups[trackKey].push(node);
      }
    });
    
    // Assign tracks for each group
    Object.values(trackGroups).forEach(group => {
      // If the group is a single node with a single parent, try to inherit the parent's track
      if (group.length === 1 && group[0].parentIds.length === 1) {
        const parentId = group[0].parentIds[0];
        if (parentId && graph[parentId]) {
          group[0].track = graph[parentId].track;
          assigned.add(group[0].id);
          return;
        }
      }
      
      // For join nodes or groups with multiple nodes, assign new tracks
      // For fork nodes, spread children across consecutive tracks
      group.forEach((node, index) => {
        if (node.parentIds.length > 1) {
          // Calculate the average track of parents for join nodes
          const parentTracks = node.parentIds
            .map(id => graph[id]?.track || 0)
            .filter(track => track !== undefined);
          
          if (parentTracks.length > 0) {
            const avgTrack = parentTracks.reduce((a, b) => a + b, 0) / parentTracks.length;
            node.track = Math.round(avgTrack);
          } else {
            node.track = currentTrack++;
          }
        } 
        // For fork children, assign consecutive tracks
        else if (node.parentIds.length === 1 && graph[node.parentIds[0]]?.childIds.length > 1) {
          const parentNode = graph[node.parentIds[0]];
          // Center the fork children around the parent track
          const childCount = parentNode.childIds.length;
          const offset = Math.floor(childCount / 2);
          node.track = parentNode.track - offset + index;
        }
        // For standalone nodes or simple sequential nodes
        else {
          node.track = currentTrack++;
        }
        
        assigned.add(node.id);
      });
    });
  });
  
  // Normalize tracks to start from 0
  const tracks = new Set<number>();
  Object.values(graph).forEach(node => tracks.add(node.track));
  const sortedTracks = Array.from(tracks).sort((a, b) => a - b);
  
  // Create a mapping from old tracks to new normalized tracks
  const trackMap: Record<number, number> = {};
  sortedTracks.forEach((track, index) => {
    trackMap[track] = index;
  });
  
  // Apply the normalized tracks
  Object.values(graph).forEach(node => {
    node.track = trackMap[node.track];
  });
  
  return tracks.size;
}

// Create a grid representation of nodes
function createNodeGrid(graph: Record<string, WorkflowNode>): WorkflowNode[][] {
  // Find the maximum level
  const maxLevel = Math.max(...Object.values(graph).map(node => node.level));
  // Find the maximum track
  const maxTrack = Math.max(...Object.values(graph).map(node => node.track));
  
  // Initialize the grid
  const grid: WorkflowNode[][] = Array(maxLevel + 1)
    .fill(null)
    .map(() => Array(maxTrack + 1).fill(null));
  
  // Place nodes in the grid
  Object.values(graph).forEach(node => {
    grid[node.level][node.track] = node;
  });
  
  return grid;
}

// Calculate line animations for connections
function calculateLineAnimations(
  grid: WorkflowNode[][],
  graph: Record<string, WorkflowNode>,
  transitions: { id: string; from: string; to: string; condition?: string; workflowId: string }[],
  completedSteps: string[]
) {
  // Track the completion percentage for horizontal and vertical lines
  const horizontalLines: number[][] = Array(grid.length)
    .fill(null)
    .map(() => Array(grid[0]?.length || 0).fill(0));
  
  const verticalLines: number[][] = Array(grid.length)
    .fill(null)
    .map(() => Array(grid[0]?.length || 0).fill(0));
  
  // Calculate connection lines between nodes
  const connectionLines: { from: string; to: string; percent: number }[] = [];
  
  transitions.forEach(transition => {
    const fromNode = graph[transition.from];
    const toNode = graph[transition.to];
    
    if (!fromNode || !toNode) return;
    
    // Calculate completion percentage for this transition
    let percent = 0;
    
    if (completedSteps.includes(fromNode.id) && completedSteps.includes(toNode.id)) {
      percent = 100; // Both nodes completed
    } else if (completedSteps.includes(fromNode.id)) {
      percent = 50; // Only source node completed
    }
    
    connectionLines.push({
      from: fromNode.id,
      to: toNode.id,
      percent
    });
    
    // Update horizontal line for the source node's level
    if (percent > 0) {
      horizontalLines[fromNode.level][fromNode.track] = percent;
    }
    
    // Update vertical line for the destination node
    if (percent > 0) {
      verticalLines[toNode.level][toNode.track] = percent;
    }
  });
  
  return {
    horizontalLines,
    verticalLines,
    connectionLines
  };
} 