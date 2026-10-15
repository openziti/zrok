/**
 * Demo Page Component
 *
 * A full page added by the demo extension, accessible at /demo.
 * Demonstrates how extensions can add complete pages to the UI.
 */

import React, { useState } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  Container,
  Grid2,
  Typography,
  AppBar,
  Toolbar,
  Chip,
} from '@mui/material';
import { useNavigate } from 'react-router';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import AddIcon from '@mui/icons-material/Add';
import RemoveIcon from '@mui/icons-material/Remove';
import SettingsIcon from '@mui/icons-material/Settings';
import CodeIcon from '@mui/icons-material/Code';
import { ExtensionRouteProps, ScriptInjector, useScriptInjector } from '../../../src/extensions';
import { DemoExtensionState } from './index';

const DemoPage: React.FC<ExtensionRouteProps> = ({ user, context, logout }) => {
  const navigate = useNavigate();

  // Get extension state
  const state = context.getState<DemoExtensionState>();
  const counter = state?.counter ?? 0;
  const lastVisited = state?.lastVisited;

  // Runtime script injection demo
  const [runtimeScriptLoaded, setRuntimeScriptLoaded] = useState(false);
  const [hookScriptLoaded, setHookScriptLoaded] = useState(false);
  const { injectScript, removeScript, isLoaded } = useScriptInjector();

  const handleIncrement = () => {
    context.setState<DemoExtensionState>({ counter: counter + 1 });
    context.notify('Counter incremented!', 'info');
  };

  const handleDecrement = () => {
    context.setState<DemoExtensionState>({ counter: counter - 1 });
    context.notify('Counter decremented!', 'info');
  };

  const handleBack = () => {
    navigate('/');
  };

  const handleSettings = () => {
    navigate('/demo/settings');
  };

  // Handler for useScriptInjector hook demo
  const handleInjectScript = async () => {
    try {
      await injectScript({
        id: 'demo-hook-injected-script',
        content: `
          console.log('[Demo Extension] Script injected via useScriptInjector hook!');
          window.demoHookInjected = true;
        `,
      });
      setHookScriptLoaded(true);
      context.notify('Script injected via hook!', 'success');
    } catch (error) {
      context.notify('Failed to inject script', 'error');
    }
  };

  const handleRemoveScript = () => {
    const removed = removeScript('demo-hook-injected-script');
    if (removed) {
      setHookScriptLoaded(false);
      context.notify('Script removed!', 'info');
    }
  };

  return (
    <Box sx={{ flexGrow: 1 }}>
      <AppBar position="static">
        <Toolbar>
          <Button color="inherit" onClick={handleBack} startIcon={<ArrowBackIcon />}>
            Back to Console
          </Button>
          <Typography variant="h6" sx={{ flexGrow: 1, ml: 2 }}>
            Demo Extension
          </Typography>
          <Button color="inherit" onClick={handleSettings} startIcon={<SettingsIcon />}>
            Settings
          </Button>
          <Button color="inherit" onClick={logout}>
            Logout
          </Button>
        </Toolbar>
      </AppBar>

      <Container maxWidth="md" sx={{ mt: 4 }}>
        <Typography variant="h4" gutterBottom>
          Demo Extension Page
        </Typography>

        <Typography variant="body1" paragraph>
          This page demonstrates how extensions can add complete new pages to the zrok UI.
          The page has access to the current user, extension context, and can interact with
          the extension's state.
        </Typography>

        <Grid2 container spacing={3}>
          <Grid2 size={{ xs: 12, md: 6 }}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  User Information
                </Typography>
                <Typography variant="body2">
                  <strong>Email:</strong> {user?.email ?? 'Not logged in'}
                </Typography>
                <Typography variant="body2">
                  <strong>Extension ID:</strong> {context.extensionId}
                </Typography>
                {lastVisited && (
                  <Typography variant="body2">
                    <strong>Last Visited:</strong> {new Date(lastVisited).toLocaleString()}
                  </Typography>
                )}
              </CardContent>
            </Card>
          </Grid2>

          <Grid2 size={{ xs: 12, md: 6 }}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  State Demo: Counter
                </Typography>
                <Typography variant="h2" align="center" sx={{ my: 2 }}>
                  {counter}
                </Typography>
                <Box sx={{ display: 'flex', justifyContent: 'center', gap: 2 }}>
                  <Button
                    variant="contained"
                    color="primary"
                    onClick={handleDecrement}
                    startIcon={<RemoveIcon />}
                  >
                    Decrement
                  </Button>
                  <Button
                    variant="contained"
                    color="primary"
                    onClick={handleIncrement}
                    startIcon={<AddIcon />}
                  >
                    Increment
                  </Button>
                </Box>
                <Typography variant="caption" display="block" sx={{ mt: 2, textAlign: 'center' }}>
                  This counter persists in the Zustand store
                </Typography>
              </CardContent>
            </Card>
          </Grid2>

          <Grid2 size={{ xs: 12 }}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  <CodeIcon sx={{ mr: 1, verticalAlign: 'middle' }} />
                  Runtime Script Injection Demo
                </Typography>
                <Typography variant="body2" paragraph>
                  Extensions can inject scripts at runtime using the <code>ScriptInjector</code> component
                  or the <code>useScriptInjector</code> hook. Check the browser console to see the output.
                </Typography>

                <Box sx={{ mb: 3 }}>
                  <Typography variant="subtitle2" gutterBottom>
                    ScriptInjector Component (Declarative)
                  </Typography>
                  <Typography variant="body2" color="text.secondary" paragraph>
                    The ScriptInjector component below is rendered when this page loads.
                    It automatically injects an inline script and removes it on unmount.
                  </Typography>
                  <Chip
                    label={runtimeScriptLoaded ? 'Script Loaded' : 'Loading...'}
                    color={runtimeScriptLoaded ? 'success' : 'default'}
                    size="small"
                  />
                  {/* Declarative script injection via component */}
                  <ScriptInjector
                    id="demo-component-injected-script"
                    content={`
                      console.log('[Demo Extension] Script injected via ScriptInjector component!');
                      window.demoComponentInjected = true;
                    `}
                    onLoad={() => setRuntimeScriptLoaded(true)}
                    removeOnUnmount={true}
                  />
                </Box>

                <Box>
                  <Typography variant="subtitle2" gutterBottom>
                    useScriptInjector Hook (Imperative)
                  </Typography>
                  <Typography variant="body2" color="text.secondary" paragraph>
                    Use the hook for programmatic control over script injection.
                  </Typography>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                    <Button
                      variant="outlined"
                      size="small"
                      onClick={handleInjectScript}
                      disabled={hookScriptLoaded}
                    >
                      Inject Script
                    </Button>
                    <Button
                      variant="outlined"
                      size="small"
                      onClick={handleRemoveScript}
                      disabled={!hookScriptLoaded}
                    >
                      Remove Script
                    </Button>
                    <Chip
                      label={hookScriptLoaded ? 'Script Loaded' : 'Not Loaded'}
                      color={hookScriptLoaded ? 'success' : 'default'}
                      size="small"
                    />
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </Grid2>

          <Grid2 size={{ xs: 12 }}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Extension Capabilities Demonstrated
                </Typography>
                <ul>
                  <li>Custom route at <code>/demo</code></li>
                  <li>Navigation item in the navbar</li>
                  <li>State management via extension context</li>
                  <li>User information access</li>
                  <li>Notification system</li>
                  <li>Navigation between pages</li>
                  <li>Build-time script injection (headScripts/bodyScripts)</li>
                  <li>Runtime script injection (ScriptInjector component)</li>
                  <li>Imperative script injection (useScriptInjector hook)</li>
                </ul>
              </CardContent>
            </Card>
          </Grid2>
        </Grid2>
      </Container>
    </Box>
  );
};

export default DemoPage;
