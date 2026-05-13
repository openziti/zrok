import React from "react";
import {Box, Button, Typography} from "@mui/material";

interface ErrorBoundaryProps {
    children: React.ReactNode;
    fallback?: React.ReactNode | ((props: { error: Error; reset: () => void }) => React.ReactNode);
}

interface ErrorBoundaryState {
    hasError: boolean;
    error: Error | null;
}

class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
    state: ErrorBoundaryState = { hasError: false, error: null };

    static getDerivedStateFromError(error: Error): ErrorBoundaryState {
        return { hasError: true, error };
    }

    reset = () => {
        this.setState({ hasError: false, error: null });
    };

    render() {
        if (this.state.hasError) {
            if (typeof this.props.fallback === "function") {
                return this.props.fallback({ error: this.state.error!, reset: this.reset });
            }
            if (this.props.fallback) {
                return this.props.fallback;
            }
            return (
                <Box sx={{ p: 3, textAlign: "center" }}>
                    <Typography color="error">Something went wrong.</Typography>
                    <Button onClick={this.reset} variant="outlined" sx={{ mt: 1 }}>Try Again</Button>
                </Box>
            );
        }
        return this.props.children;
    }
}

export default ErrorBoundary;
