/**
 * Extension Context Factory
 *
 * Creates context objects that extensions use to interact with the zrok UI.
 * Each extension gets its own context with access to state management,
 * navigation, notifications, and subscriptions.
 */

import { ExtensionContext } from './types';
import useApiConsoleStore from '../model/store';

/**
 * Create an extension context for a specific extension.
 *
 * @param extensionId - The unique ID of the extension
 * @param navigate - Navigation function (from react-router)
 * @param notify - Notification function
 */
export function createExtensionContext(
  extensionId: string,
  navigate: (path: string) => void,
  notify: (message: string, severity?: 'info' | 'success' | 'warning' | 'error') => void
): ExtensionContext {
  return {
    extensionId,

    getState: <T = Record<string, unknown>>(): T | undefined => {
      const state = useApiConsoleStore.getState();
      return state.extensions?.[extensionId] as T | undefined;
    },

    setState: <T = Record<string, unknown>>(partialState: Partial<T>): void => {
      const { setExtensionState } = useApiConsoleStore.getState();
      setExtensionState(extensionId, partialState);
    },

    subscribe: <T = Record<string, unknown>>(
      selector: (state: T) => unknown,
      callback: (selectedValue: unknown, previousValue: unknown) => void
    ): (() => void) => {
      let previousValue: unknown;

      return useApiConsoleStore.subscribe((state) => {
        const extensionState = state.extensions?.[extensionId] as T | undefined;
        if (!extensionState) return;

        const selectedValue = selector(extensionState);
        if (selectedValue !== previousValue) {
          const prevValue = previousValue;
          previousValue = selectedValue;
          callback(selectedValue, prevValue);
        }
      });
    },

    getUser: () => {
      return useApiConsoleStore.getState().user;
    },

    subscribeToUser: (callback: (user: any) => void): (() => void) => {
      let previousUser = useApiConsoleStore.getState().user;

      return useApiConsoleStore.subscribe((state) => {
        if (state.user !== previousUser) {
          previousUser = state.user;
          callback(state.user);
        }
      });
    },

    getSelectedNode: () => {
      return useApiConsoleStore.getState().selectedNode;
    },

    subscribeToSelectedNode: (callback: (node: any) => void): (() => void) => {
      let previousNode = useApiConsoleStore.getState().selectedNode;

      return useApiConsoleStore.subscribe((state) => {
        if (state.selectedNode !== previousNode) {
          previousNode = state.selectedNode;
          callback(state.selectedNode);
        }
      });
    },

    navigate,

    notify,
  };
}

/**
 * Hook for extensions to access their context within React components.
 * Must be used within a component that has access to the extension context.
 *
 * @example
 * ```tsx
 * // In an extension component
 * function MyExtensionPanel({ context }: PanelExtensionProps) {
 *   const { getState, setState } = context;
 *
 *   const handleClick = () => {
 *     setState({ clicked: true });
 *   };
 *
 *   return <button onClick={handleClick}>Click me</button>;
 * }
 * ```
 */
export function useExtensionState<T = Record<string, unknown>>(
  extensionId: string
): {
  state: T | undefined;
  setState: (partial: Partial<T>) => void;
} {
  const state = useApiConsoleStore(
    (s) => s.extensions?.[extensionId] as T | undefined
  );
  const setExtensionState = useApiConsoleStore((s) => s.setExtensionState);

  return {
    state,
    setState: (partial: Partial<T>) => setExtensionState(extensionId, partial),
  };
}
