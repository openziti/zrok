/**
 * Demo Icon Component
 *
 * A simple icon component for the demo extension.
 */

import React from 'react';
import ScienceIcon from '@mui/icons-material/Science';

interface DemoIconProps {
  fontSize?: 'small' | 'medium' | 'large';
}

const DemoIcon: React.FC<DemoIconProps> = ({ fontSize = 'medium' }) => {
  return <ScienceIcon fontSize={fontSize} />;
};

export default DemoIcon;
