import React from 'react';
import styled from 'styled-components';

interface CheckboxProps {
  name?: string;
  checked?: boolean;
  onChange?: (e: React.ChangeEvent<HTMLInputElement>) => void;
  darkMode?: boolean;
  className?: string;
}

const Checkbox: React.FC<CheckboxProps> = ({ 
  name,
  checked, 
  onChange, 
  darkMode = false,
  className,
  ...props 
}) => {
  return (
    <StyledWrapper $darkMode={darkMode}>
      <input
        type="checkbox"
        name={name}
        className={`ui-checkbox ${className || ''}`}
        checked={checked}
        onChange={onChange}
        {...props}
      />
    </StyledWrapper>
  );
};

Checkbox.displayName = 'Checkbox';

const StyledWrapper = styled.div<{ $darkMode: boolean }>`
  display: flex;
  align-items: center;
  gap: 8px;

  /* checkbox settings */
  .ui-checkbox {
    --primary-color: ${props => props.$darkMode ? '#0A84FF' : '#1677ff'};
    --secondary-color: ${props => props.$darkMode ? 'hsl(220, 3%, 13%)' : '#fff'};
    --primary-hover-color: ${props => props.$darkMode ? '#409cff' : '#4096ff'};
    /* checkbox */
    --checkbox-diameter: 16px;
    --checkbox-border-radius: 4px;
    --checkbox-border-color: ${props => props.$darkMode ? 'hsl(220, 3%, 18%)' : 'hsl(214.3, 31.8%, 91.4%)'};
    --checkbox-border-width: 1px;
    --checkbox-border-style: solid;
    --checkmark-size: 1.2;
  }

  .ui-checkbox,
  .ui-checkbox *,
  .ui-checkbox *::before,
  .ui-checkbox *::after {
    -webkit-box-sizing: border-box;
    box-sizing: border-box;
  }

  .ui-checkbox {
    -webkit-appearance: none;
    -moz-appearance: none;
    appearance: none;
    width: var(--checkbox-diameter);
    height: var(--checkbox-diameter);
    border-radius: var(--checkbox-border-radius);
    background: var(--secondary-color);
    border: var(--checkbox-border-width) var(--checkbox-border-style) var(--checkbox-border-color);
    -webkit-transition: all 0.3s;
    -o-transition: all 0.3s;
    transition: all 0.3s;
    cursor: pointer;
    position: relative;
  }

  .ui-checkbox::after {
    content: "";
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    -webkit-box-shadow: 0 0 0 calc(var(--checkbox-diameter) / 2.5) var(--primary-color);
    box-shadow: 0 0 0 calc(var(--checkbox-diameter) / 2.5) var(--primary-color);
    border-radius: inherit;
    opacity: 0;
    -webkit-transition: all 0.5s cubic-bezier(0.12, 0.4, 0.29, 1.46);
    -o-transition: all 0.5s cubic-bezier(0.12, 0.4, 0.29, 1.46);
    transition: all 0.5s cubic-bezier(0.12, 0.4, 0.29, 1.46);
  }

  .ui-checkbox::before {
    top: 45%;
    left: 50%;
    content: "";
    position: absolute;
    width: 4px;
    height: 7px;
    border-right: 2px solid #fff;
    border-bottom: 2px solid #fff;
    -webkit-transform: translate(-50%, -50%) rotate(45deg) scale(0);
    -ms-transform: translate(-50%, -50%) rotate(45deg) scale(0);
    transform: translate(-50%, -50%) rotate(45deg) scale(0);
    opacity: 0;
    -webkit-transition: all 0.1s cubic-bezier(0.71, -0.46, 0.88, 0.6), opacity 0.1s;
    -o-transition: all 0.1s cubic-bezier(0.71, -0.46, 0.88, 0.6), opacity 0.1s;
    transition: all 0.1s cubic-bezier(0.71, -0.46, 0.88, 0.6), opacity 0.1s;
  }

  /* actions */
  .ui-checkbox:hover {
    border-color: var(--primary-color);
    background: ${props => props.$darkMode ? 'hsl(220, 3%, 16%)' : 'var(--secondary-color)'};
  }

  .ui-checkbox:checked {
    background: var(--primary-color);
    border-color: transparent;
  }

  .ui-checkbox:checked:hover {
    background: var(--primary-hover-color);
  }

  .ui-checkbox:checked::before {
    opacity: 1;
    -webkit-transform: translate(-50%, -50%) rotate(45deg) scale(var(--checkmark-size));
    -ms-transform: translate(-50%, -50%) rotate(45deg) scale(var(--checkmark-size));
    transform: translate(-50%, -50%) rotate(45deg) scale(var(--checkmark-size));
    -webkit-transition: all 0.2s cubic-bezier(0.12, 0.4, 0.29, 1.46) 0.1s;
    -o-transition: all 0.2s cubic-bezier(0.12, 0.4, 0.29, 1.46) 0.1s;
    transition: all 0.2s cubic-bezier(0.12, 0.4, 0.29, 1.46) 0.1s;
  }

  .ui-checkbox:active:not(:checked)::after {
    -webkit-transition: none;
    -o-transition: none;
    -webkit-box-shadow: none;
    box-shadow: none;
    transition: none;
    opacity: 1;
  }

  /* Label styling */
  .checkbox-label {
    font-size: 14px;
    color: ${(props) => (props.$darkMode ? '#fff' : '#000')};
    user-select: none;
  }
`;

export default Checkbox;