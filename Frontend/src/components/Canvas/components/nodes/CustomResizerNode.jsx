import { Handle, Position, NodeResizer } from '@xyflow/react';
import { useState, useCallback } from 'react';

const CustomResizerNode = ({ id, data, selected, isConnectable = true, updateNodeLabel }) => {
  const [isEditing, setIsEditing] = useState(false);
  const [labelValue, setLabelValue] = useState(data.label);
  const [nodeDimensions, setNodeDimensions] = useState({
    width: undefined,
    height: undefined,
  });

  const onDoubleClick = useCallback(() => {
    setIsEditing(true);
    setLabelValue(data.label);
  }, [data.label]);

  const onBlur = useCallback(() => {
    setIsEditing(false);
    if (labelValue !== data.label) {
      updateNodeLabel(id, labelValue);
    }
  }, [id, labelValue, data.label, updateNodeLabel]);

  const onChange = useCallback((evt) => {
    setLabelValue(evt.target.value);
  }, []);

  const onKeyDown = useCallback((evt) => {
    if (evt.key === 'Enter') {
      evt.preventDefault();
      evt.target.blur();
    }
    if (evt.key === 'Escape') {
      setLabelValue(data.label);
      setIsEditing(false);
    }
  }, [data.label]);

  return (
    <div style={{ height: '100%', width: '100%' }}>
      {/* Show the NodeResizer only when the node is selected */}
      {selected && (
        <NodeResizer
          minWidth={100}
          minHeight={50}
          onResize={(event, { width, height }) => {
            setNodeDimensions({ width, height });
          }}
          isVisible={selected}
        />
      )}

      <div className="flex flex-col h-full justify-between p-4">
        {isEditing ? (
          <input
            className="nodrag text-base font-medium mb-2 bg-transparent border-none outline-none focus:outline-none"
            value={labelValue}
            onChange={onChange}
            onBlur={onBlur}
            onKeyDown={onKeyDown}
            autoFocus
          />
        ) : (
          <div 
            className="text-base font-medium mb-2"
            onDoubleClick={onDoubleClick}
          >
            {data.label}
          </div>
        )}
        {data.content && <div className="text-sm">{data.content}</div>}
      </div>

      {/* Add handles to allow connections */}
      <Handle
        id="top"
        type="target"
        position={Position.Top}
        className="w-2 h-2 !bg-primary"
        isConnectable={isConnectable}
      />
      <Handle
        id="right"
        type="source"
        position={Position.Right}
        className="w-2 h-2 !bg-primary"
        isConnectable={isConnectable}
      />
      <Handle
        id="bottom"
        type="source"
        position={Position.Bottom}
        className="w-2 h-2 !bg-primary"
        isConnectable={isConnectable}
      />
      <Handle
        id="left"
        type="target"
        position={Position.Left}
        className="w-2 h-2 !bg-primary"
        isConnectable={isConnectable}
      />
    </div>
  );
};

export default CustomResizerNode;