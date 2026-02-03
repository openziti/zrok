import * as React from "react";
import {AgentObject} from "./model/overview.ts";
import {ShareDetail} from "./api";
import {
    Accordion,
    AccordionDetails,
    AccordionSummary,
    AppBar,
    Box,
    Button,
    Card,
    Chip,
    Grid2,
    Toolbar,
    Typography
} from "@mui/material";
import ShareIcon from "@mui/icons-material/Share";
import DeleteIcon from "@mui/icons-material/Delete";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import WarningIcon from "@mui/icons-material/Warning";
import ErrorIcon from "@mui/icons-material/Error";
import {GetAgentApi} from "./model/api.ts";

interface ShareCardProps {
    shareObject: AgentObject;
}

const ShareCard = ({ shareObject }: ShareCardProps) => {
    const frontends = new Array<React.JSX.Element>();
    const share = (shareObject.v as ShareDetail);

    if (share.frontendEndpoint) {
        share.frontendEndpoint.map((fe, index) => {
            frontends.push(<a key={`${shareObject.displayToken}-${index}`} href={fe} target="_blank" rel="noopener noreferrer">{fe}</a>);
        });
    }

    const getStatusColor = (status: string) => {
        switch (status) {
            case 'active': return '#4caf50'; // green
            case 'retrying': return '#ff9800'; // orange
            case 'failed': return '#f44336'; // red
            default: return '#9e9e9e'; // gray
        }
    };

    const getStatusIcon = (status: string) => {
        switch (status) {
            case 'active': return <CheckCircleIcon fontSize="small" />;
            case 'retrying': return <WarningIcon fontSize="small" />;
            case 'failed': return <ErrorIcon fontSize="small" />;
            default: return null;
        }
    };

    const formatTime = (date?: Date): string => {
        if (!date) return '-';
        const now = new Date();
        const diff = Math.floor((date.getTime() - now.getTime()) / 1000);
        if (diff > 0) {
            const hours = Math.floor(diff / 3600);
            const minutes = Math.floor((diff % 3600) / 60);
            const seconds = diff % 60;
            if (hours > 0) return `${hours}h ${minutes}m`;
            if (minutes > 0) return `${minutes}m ${seconds}s`;
            return `${seconds}s`;
        }
        return date.toLocaleTimeString();
    };

    const releaseShare = () => {
        // use the share token if available, otherwise use the failure ID
        const tokenToUse = share.token || share.failure?.id;
        if (tokenToUse) {
            GetAgentApi().agentReleaseShare({token: tokenToUse})
                .catch(e => {
                    console.log(e);
                });
        }
    }

    const cardStyle = {
        borderLeft: `4px solid ${getStatusColor(shareObject.status)}`
    };

    return (
        <Card sx={cardStyle}>
            <AppBar position="sticky">
                <Toolbar variant="dense">
                    <ShareIcon />
                    <Grid2 container sx={{ flexGrow: 1, alignItems: 'center' }}>
                        <Grid2 display="flex" justifyContent="center" size="grow">
                            <Typography variant="h6" component="div" style={{ color: "#9bf316" }}>
                                {shareObject.displayToken}
                            </Typography>
                        </Grid2>
                    </Grid2>
                    <Grid2 container sx={{ alignItems: 'center' }}>
                        <Grid2 display="flex" justifyContent="right" sx={{ alignItems: 'center' }}>
                            <Chip
                                icon={getStatusIcon(shareObject.status)}
                                label={shareObject.status}
                                size="small"
                                sx={{
                                    backgroundColor: getStatusColor(shareObject.status),
                                    color: 'white',
                                    mr: 1
                                }}
                            />
                            <Chip label={share.shareMode} size="small" style={{ backgroundColor: "#9bf316" }} sx={{ mr: 1}} />
                            <Chip label={share.backendMode} size="small" style={{ backgroundColor: "#9bf316" }} />
                        </Grid2>
                    </Grid2>
                </Toolbar>
            </AppBar>
            <Box sx={{ p: 2, textAlign: "center" }}>
                <Typography variant="h6" component="div">
                    {share.backendEndpoint} &rarr; {frontends.length > 0 ? frontends : 'no frontend endpoints'} <br/>
                </Typography>
            </Box>

            {share.failure && (
                <Accordion>
                    <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                        <Typography color="error">failure details</Typography>
                    </AccordionSummary>
                    <AccordionDetails>
                        <Typography variant="body2" gutterBottom>
                            <strong>failure count:</strong> {share.failure.count || 0}
                        </Typography>
                        {share.failure.lastError && (
                            <Typography variant="body2" gutterBottom sx={{ wordBreak: 'break-word' }}>
                                <strong>last error:</strong> {share.failure.lastError}
                            </Typography>
                        )}
                        <Typography variant="body2">
                            <strong>next retry:</strong> {formatTime(share.failure.nextRetry)}
                        </Typography>
                    </AccordionDetails>
                </Accordion>
            )}

            <Grid2 container sx={{ flexGrow: 1, p: 1 }}>
                <Grid2 display="flex" justifyContent="right" size="grow">
                    <Button variant="contained" onClick={releaseShare}><DeleteIcon /></Button>
                </Grid2>
            </Grid2>
        </Card>
    );
}

export default ShareCard;