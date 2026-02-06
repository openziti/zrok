/**
 * SlotRenderer Component
 *
 * Renders all extension components registered for a specific slot.
 * Slots are named injection points in the UI where extensions can add content.
 */

import React from 'react';
import { Node } from '@xyflow/react';
import { extensionRegistry } from './registry';
import { SlotProps, SlotName } from './types';
import { User } from '../model/user';

interface SlotRendererProps {
  /** The name of the slot to render */
  name: SlotName | string;

  /** Current user (optional) */
  user?: User | null;

  /** Currently selected node (optional) */
  selectedNode?: Node | null;

  /** Additional props to pass to slot components */
  [key: string]: unknown;
}

/**
 * Renders all components registered for a given slot.
 *
 * @example
 * ```tsx
 * // In NavBar.tsx
 * <Slot name={SLOTS.NAVBAR_RIGHT} user={user} />
 *
 * // In AccountPanel.tsx
 * <Slot name={SLOTS.ACCOUNT_PANEL_ACTIONS} user={user} selectedNode={node} />
 * ```
 */
export const Slot: React.FC<SlotRendererProps> = ({
  name,
  user,
  selectedNode,
  ...additionalProps
}) => {
  const components = extensionRegistry.getSlotComponents(name);

  if (components.length === 0) {
    return null;
  }

  return (
    <>
      {components.map(({ component: Component, extensionId }, index) => {
        const context = extensionRegistry.getContext(extensionId);

        if (!context) {
          console.warn(
            `[Slot] Extension "${extensionId}" context not found for slot "${name}"`
          );
          return null;
        }

        const slotProps: SlotProps = {
          user,
          selectedNode,
          context,
          ...additionalProps,
        };

        return (
          <SlotErrorBoundary
            key={`${name}-${extensionId}-${index}`}
            extensionId={extensionId}
            slotName={name}
          >
            <Component {...slotProps} />
          </SlotErrorBoundary>
        );
      })}
    </>
  );
};

/**
 * Error boundary for slot components.
 * Prevents one extension's error from breaking the entire UI.
 */
interface SlotErrorBoundaryProps {
  extensionId: string;
  slotName: string;
  children: React.ReactNode;
}

interface SlotErrorBoundaryState {
  hasError: boolean;
  error?: Error;
}

class SlotErrorBoundary extends React.Component<
  SlotErrorBoundaryProps,
  SlotErrorBoundaryState
> {
  constructor(props: SlotErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): SlotErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo): void {
    console.error(
      `[Slot] Error in extension "${this.props.extensionId}" slot "${this.props.slotName}":`,
      error,
      errorInfo
    );
  }

  render(): React.ReactNode {
    if (this.state.hasError) {
      // Return null to silently fail - don't break the UI
      // In development, you might want to show an error indicator
      if (import.meta.env.DEV) {
        return (
          <div
            style={{
              padding: '4px 8px',
              backgroundColor: '#ffebee',
              color: '#c62828',
              fontSize: '12px',
              borderRadius: '4px',
            }}
          >
            Extension error: {this.props.extensionId}
          </div>
        );
      }
      return null;
    }

    return this.props.children;
  }
}

export default Slot;
