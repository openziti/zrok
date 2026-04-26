/**
 * Demo Settings Page
 *
 * A settings page for the demo extension.
 * Demonstrates nested routes and form handling in extensions.
 */

import React from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  Container,
  FormControlLabel,
  Switch,
  Typography,
  AppBar,
  Toolbar,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
} from '@mui/material';
import { useNavigate } from 'react-router';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import { ExtensionRouteProps } from '../../../src/extensions';
import { DemoExtensionState } from './index';

const DemoSettingsPage: React.FC<ExtensionRouteProps> = ({ context }) => {
  const navigate = useNavigate();

  // Get extension state
  const state = context.getState<DemoExtensionState>();
  const settings = state?.settings ?? { enableFeatureX: true, theme: 'light' };

  const handleBack = () => {
    navigate('/demo');
  };

  const handleFeatureXToggle = (event: React.ChangeEvent<HTMLInputElement>) => {
    context.setState<DemoExtensionState>({
      settings: {
        ...settings,
        enableFeatureX: event.target.checked,
      },
    });
    context.notify(
      `Feature X ${event.target.checked ? 'enabled' : 'disabled'}`,
      'info'
    );
  };

  const handleThemeChange = (event: any) => {
    context.setState<DemoExtensionState>({
      settings: {
        ...settings,
        theme: event.target.value,
      },
    });
    context.notify(`Theme changed to ${event.target.value}`, 'info');
  };

  return (
    <Box sx={{ flexGrow: 1 }}>
      <AppBar position="static">
        <Toolbar>
          <Button color="inherit" onClick={handleBack} startIcon={<ArrowBackIcon />}>
            Back to Demo
          </Button>
          <Typography variant="h6" sx={{ flexGrow: 1, ml: 2 }}>
            Demo Extension Settings
          </Typography>
        </Toolbar>
      </AppBar>

      <Container maxWidth="sm" sx={{ mt: 4 }}>
        <Typography variant="h4" gutterBottom>
          Extension Settings
        </Typography>

        <Card sx={{ mt: 3 }}>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              Feature Toggles
            </Typography>

            <FormControlLabel
              control={
                <Switch
                  checked={settings.enableFeatureX}
                  onChange={handleFeatureXToggle}
                />
              }
              label="Enable Feature X"
            />

            <Typography variant="body2" color="text.secondary" sx={{ mt: 1, mb: 3 }}>
              This setting demonstrates how extensions can persist settings in the store.
            </Typography>

            <FormControl fullWidth sx={{ mt: 2 }}>
              <InputLabel id="theme-select-label">Theme</InputLabel>
              <Select
                labelId="theme-select-label"
                value={settings.theme}
                label="Theme"
                onChange={handleThemeChange}
              >
                <MenuItem value="light">Light</MenuItem>
                <MenuItem value="dark">Dark</MenuItem>
              </Select>
            </FormControl>

            <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
              Note: This is a demo setting and does not actually change the theme.
            </Typography>
          </CardContent>
        </Card>

        <Card sx={{ mt: 3 }}>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              Current State
            </Typography>
            <pre style={{ fontSize: '12px', overflow: 'auto' }}>
              {JSON.stringify(state, null, 2)}
            </pre>
          </CardContent>
        </Card>
      </Container>
    </Box>
  );
};

export default DemoSettingsPage;
