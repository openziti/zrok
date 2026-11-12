/**
 * Demo Navbar Slot Component
 *
 * A component injected into the NAVBAR_RIGHT slot.
 * Demonstrates how extensions can inject UI into predefined slots.
 */

import React from 'react';
import { Badge, Button, Tooltip } from '@mui/material';
import NotificationsIcon from '@mui/icons-material/Notifications';
import { SlotProps } from '../../../src/extensions';
import { DemoExtensionState } from './index';

const DemoNavbarSlot: React.FC<SlotProps> = ({ user, context }) => {
  // Get the counter from extension state to show as a badge
  const state = context.getState<DemoExtensionState>();
  const counter = state?.counter ?? 0;

  const handleClick = () => {
    context.notify(`You have ${counter} notifications (demo)`, 'info');
  };

  // Only show if user is logged in
  if (!user) return null;

  return (
    <Tooltip title="Demo Notifications">
      <Button color="inherit" onClick={handleClick}>
        <Badge badgeContent={counter} color="error" max={99}>
          <NotificationsIcon />
        </Badge>
      </Button>
    </Tooltip>
  );
};

export default DemoNavbarSlot;
