/**
 * PanelWrapper Component
 *
 * Wraps the built-in panels (AccountPanel, EnvironmentPanel, etc.) and
 * injects extension panel components based on their position (before, after, tab, replace).
 */

import React, {useState, useMemo} from 'react';
import {Box, Tab, Tabs} from '@mui/material';
import {Node} from '@xyflow/react';
import {extensionRegistry} from './registry';
import {PanelExtension, PanelExtensionProps, SLOTS} from './types';
import {Slot} from './SlotRenderer';
import useApiConsoleStore from '../model/store';

interface PanelWrapperProps {
  /** The node type (account, environment, share, access) */
  nodeType: string;
  /** The selected node */
  node: Node;
  /** The built-in panel component */
  children: React.ReactNode;
}

interface TabPanelProps {
  children?: React.ReactNode;
  value: number;
  index: number;
}

const TabPanel: React.FC<TabPanelProps> = ({ children, value, index }) => {
  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`panel-tabpanel-${index}`}
      aria-labelledby={`panel-tab-${index}`}
    >
      {value === index && <Box sx={{ pt: 2 }}>{children}</Box>}
    </div>
  );
};

export const PanelWrapper: React.FC<PanelWrapperProps> = ({
  nodeType,
  node,
  children,
}) => {
  const user = useApiConsoleStore((state) => state.user);
  const [activeTab, setActiveTab] = useState(0);

  // Get panel extensions for this node type
  const beforeExtensions = useMemo(
    () => extensionRegistry.getPanelExtensions(nodeType, 'before'),
    [nodeType]
  );
  const afterExtensions = useMemo(
    () => extensionRegistry.getPanelExtensions(nodeType, 'after'),
    [nodeType]
  );
  const tabExtensions = useMemo(
    () => extensionRegistry.getPanelExtensions(nodeType, 'tab'),
    [nodeType]
  );
  const replaceExtensions = useMemo(
    () => extensionRegistry.getPanelExtensions(nodeType, 'replace'),
    [nodeType]
  );

  // Get the appropriate slot name for this panel type
  const getSlotNames = () => {
    switch (nodeType) {
      case 'account':
        return {
          top: SLOTS.ACCOUNT_PANEL_TOP,
          bottom: SLOTS.ACCOUNT_PANEL_BOTTOM,
          actions: SLOTS.ACCOUNT_PANEL_ACTIONS,
        };
      case 'environment':
        return {
          top: SLOTS.ENVIRONMENT_PANEL_TOP,
          bottom: SLOTS.ENVIRONMENT_PANEL_BOTTOM,
        };
      case 'share':
        return {
          top: SLOTS.SHARE_PANEL_TOP,
          bottom: SLOTS.SHARE_PANEL_BOTTOM,
        };
      default:
        return {};
    }
  };

  const slots = getSlotNames();

  // Render extension panel component
  const renderExtension = (
    ext: PanelExtension & { extensionId: string },
    index: number
  ) => {
    const context = extensionRegistry.getContext(ext.extensionId);
    if (!context) return null;

    const Component = ext.component;
    const props: PanelExtensionProps = {
      node,
      user,
      context,
    };

    return (
      <ExtensionErrorBoundary
        key={`${ext.extensionId}-${index}`}
        extensionId={ext.extensionId}
      >
        <Component {...props} />
      </ExtensionErrorBoundary>
    );
  };

  // If there's a replace extension, use it instead of the built-in panel
  if (replaceExtensions.length > 0) {
    const ext = replaceExtensions[0]; // Only use the first replace extension
    return (
      <>
        {/* Slots still work with replaced panels */}
        {slots.top && <Slot name={slots.top} user={user} selectedNode={node} />}
        {renderExtension(ext, 0)}
        {slots.bottom && <Slot name={slots.bottom} user={user} selectedNode={node} />}
      </>
    );
  }

  // If there are tab extensions, render as tabbed interface
  if (tabExtensions.length > 0) {
    const handleTabChange = (_: React.SyntheticEvent, newValue: number) => {
      setActiveTab(newValue);
    };

    return (
      <Box sx={{ width: '100%' }}>
        {/* Slots at top */}
        {slots.top && <Slot name={slots.top} user={user} selectedNode={node} />}

        {/* Before extensions */}
        {beforeExtensions.map(renderExtension)}

        {/* Tabs */}
        <Tabs
          value={activeTab}
          onChange={handleTabChange}
          aria-label="panel tabs"
          sx={{ borderBottom: 1, borderColor: 'divider' }}
        >
          <Tab label="Details" id="panel-tab-0" aria-controls="panel-tabpanel-0" />
          {tabExtensions.map((ext, index) => (
            <Tab
              key={`tab-${ext.extensionId}-${index}`}
              label={ext.tabLabel || ext.extensionId}
              id={`panel-tab-${index + 1}`}
              aria-controls={`panel-tabpanel-${index + 1}`}
              icon={ext.tabIcon ? <ext.tabIcon /> : undefined}
              iconPosition="start"
            />
          ))}
        </Tabs>

        {/* Tab panels */}
        <TabPanel value={activeTab} index={0}>
          {children}
        </TabPanel>
        {tabExtensions.map((ext, index) => (
          <TabPanel key={`tabpanel-${ext.extensionId}-${index}`} value={activeTab} index={index + 1}>
            {renderExtension(ext, index)}
          </TabPanel>
        ))}

        {/* After extensions */}
        {afterExtensions.map(renderExtension)}

        {/* Slots at bottom */}
        {slots.bottom && <Slot name={slots.bottom} user={user} selectedNode={node} />}
      </Box>
    );
  }

  // Default: render with before/after extensions
  return (
    <>
      {/* Slots at top */}
      {slots.top && <Slot name={slots.top} user={user} selectedNode={node} />}

      {/* Before extensions */}
      {beforeExtensions.map(renderExtension)}

      {/* Built-in panel */}
      {children}

      {/* After extensions */}
      {afterExtensions.map(renderExtension)}

      {/* Slots at bottom */}
      {slots.bottom && <Slot name={slots.bottom} user={user} selectedNode={node} />}
    </>
  );
};

/**
 * Error boundary for extension panel components.
 */
interface ExtensionErrorBoundaryProps {
  extensionId: string;
  children: React.ReactNode;
}

interface ExtensionErrorBoundaryState {
  hasError: boolean;
  error?: Error;
}

class ExtensionErrorBoundary extends React.Component<
  ExtensionErrorBoundaryProps,
  ExtensionErrorBoundaryState
> {
  constructor(props: ExtensionErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): ExtensionErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo): void {
    console.error(
      `[PanelWrapper] Error in extension "${this.props.extensionId}":`,
      error,
      errorInfo
    );
  }

  render(): React.ReactNode {
    if (this.state.hasError) {
      if (import.meta.env.DEV) {
        return (
          <Box
            sx={{
              p: 2,
              backgroundColor: '#ffebee',
              color: '#c62828',
              borderRadius: 1,
              my: 1,
            }}
          >
            Extension error: {this.props.extensionId}
            <br />
            <small>{this.state.error?.message}</small>
          </Box>
        );
      }
      return null;
    }

    return this.props.children;
  }
}

export default PanelWrapper;
