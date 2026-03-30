/**
 * Account Billing Tab
 *
 * A tab added to the Account panel demonstrating panel extensions.
 * In a real billing extension, this would show subscription info, invoices, etc.
 */

import React, { useState } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  Divider,
  List,
  ListItem,
  ListItemText,
  Typography,
} from '@mui/material';
import CreditCardIcon from '@mui/icons-material/CreditCard';
import ReceiptIcon from '@mui/icons-material/Receipt';
import { PanelExtensionProps } from '../../../src/extensions';

// Mock data for demonstration
const mockSubscription = {
  plan: 'Professional',
  status: 'active',
  nextBillingDate: '2024-02-15',
  amount: '$49.00/month',
};

const mockInvoices = [
  { id: 'INV-001', date: '2024-01-15', amount: '$49.00', status: 'paid' },
  { id: 'INV-002', date: '2023-12-15', amount: '$49.00', status: 'paid' },
  { id: 'INV-003', date: '2023-11-15', amount: '$49.00', status: 'paid' },
];

const AccountBillingTab: React.FC<PanelExtensionProps> = ({ node, user, context }) => {
  const [loading, setLoading] = useState(false);

  const handleManageSubscription = () => {
    context.notify('Opening subscription management...', 'info');
    // In a real extension, this would open a modal or navigate to a billing page
    context.navigate('/demo');
  };

  const handleViewInvoice = (invoiceId: string) => {
    context.notify(`Viewing invoice ${invoiceId}`, 'info');
  };

  return (
    <Box sx={{ p: 1 }}>
      <Typography variant="h6" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
        <CreditCardIcon />
        Billing & Subscription
      </Typography>

      <Card variant="outlined" sx={{ mb: 2 }}>
        <CardContent>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
            <Typography variant="subtitle1" fontWeight="bold">
              Current Plan
            </Typography>
            <Chip
              label={mockSubscription.status}
              color={mockSubscription.status === 'active' ? 'success' : 'default'}
              size="small"
            />
          </Box>

          <Typography variant="h5" color="primary" gutterBottom>
            {mockSubscription.plan}
          </Typography>

          <Typography variant="body2" color="text.secondary">
            {mockSubscription.amount}
          </Typography>

          <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
            Next billing: {mockSubscription.nextBillingDate}
          </Typography>

          <Button
            variant="outlined"
            size="small"
            sx={{ mt: 2 }}
            onClick={handleManageSubscription}
          >
            Manage Subscription
          </Button>
        </CardContent>
      </Card>

      <Typography variant="subtitle1" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
        <ReceiptIcon />
        Recent Invoices
      </Typography>

      <List dense>
        {mockInvoices.map((invoice, index) => (
          <React.Fragment key={invoice.id}>
            <ListItem
              secondaryAction={
                <Button size="small" onClick={() => handleViewInvoice(invoice.id)}>
                  View
                </Button>
              }
            >
              <ListItemText
                primary={invoice.id}
                secondary={`${invoice.date} - ${invoice.amount}`}
              />
              <Chip
                label={invoice.status}
                color={invoice.status === 'paid' ? 'success' : 'warning'}
                size="small"
                sx={{ mr: 1 }}
              />
            </ListItem>
            {index < mockInvoices.length - 1 && <Divider />}
          </React.Fragment>
        ))}
      </List>

      <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mt: 2 }}>
        This is a demo billing tab. In a real extension, this would connect to your billing system.
      </Typography>
    </Box>
  );
};

export default AccountBillingTab;
