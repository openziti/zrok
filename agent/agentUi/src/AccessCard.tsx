import {AgentObject} from "./model/overview.ts";
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
import LanIcon from "@mui/icons-material/Lan";
import {AccessDetail} from "./api";
import DeleteIcon from "@mui/icons-material/Delete";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import WarningIcon from "@mui/icons-material/Warning";
import ErrorIcon from "@mui/icons-material/Error";
import {GetAgentApi} from "./model/api.ts";

interface AccessCardProps {
    accessObject: AgentObject;
}

const AccessCard = ({ accessObject }: AccessCardProps) => {
    const access = (accessObject.v as AccessDetail);

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

    const releaseAccess = () => {
        // use the frontend token if available, otherwise use the failure ID
        const tokenToUse = access.frontendToken || access.failure?.id;
        if (tokenToUse) {
            GetAgentApi().agentReleaseAccess({frontendToken: tokenToUse})
                .catch(e => {
                    console.log("error releasing access", e);
                });
        }
    }

    const cardStyle = {
        borderLeft: `4px solid ${getStatusColor(accessObject.status)}`
    };

    return (
        <Card sx={cardStyle}>
            <AppBar position="sticky">
                <Toolbar variant="dense">
                    <LanIcon />
                    <Grid2 container sx={{ flexGrow: 1, alignItems: 'center' }}>
                        <Grid2 display="flex" justifyContent="center" size="grow">
                            <Typography variant="h6" component="div" style={{ color: "#9bf316" }}>
                                {accessObject.displayToken}
                            </Typography>
                        </Grid2>
                    </Grid2>
                    <Grid2 container sx={{ alignItems: 'center' }}>
                        <Grid2 display="flex" justifyContent="right" sx={{ alignItems: 'center' }}>
                            <Chip
                                icon={getStatusIcon(accessObject.status)}
                                label={accessObject.status}
                                size="small"
                                sx={{
                                    backgroundColor: getStatusColor(accessObject.status),
                                    color: 'white',
                                    mr: 1
                                }}
                            />
                            <Chip label="private" size="small" style={{ backgroundColor: "#9bf316" }} />
                        </Grid2>
                    </Grid2>
                </Toolbar>
            </AppBar>
            <Box sx={{ p: 2, textAlign: "center" }}>
                <Typography variant="h6" component="div">
                    {access.token || 'no access token'} &rarr; {access.bindAddress || 'no bind address'}
                </Typography>
            </Box>

            {access.failure && (
                <Accordion>
                    <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                        <Typography color="error">failure details</Typography>
                    </AccordionSummary>
                    <AccordionDetails>
                        <Typography variant="body2" gutterBottom>
                            <strong>failure count:</strong> {access.failure.count || 0}
                        </Typography>
                        {access.failure.lastError && (
                            <Typography variant="body2" gutterBottom sx={{ wordBreak: 'break-word' }}>
                                <strong>last error:</strong> {access.failure.lastError}
                            </Typography>
                        )}
                        <Typography variant="body2">
                            <strong>next retry:</strong> {formatTime(access.failure.nextRetry)}
                        </Typography>
                    </AccordionDetails>
                </Accordion>
            )}

            <Grid2 container sx={{ flexGrow: 1, p: 1 }}>
                <Grid2 display="flex" justifyContent="right" size="grow">
                    <Button variant="contained" onClick={releaseAccess}><DeleteIcon /></Button>
                </Grid2>
            </Grid2>
        </Card>
    );
}

export default AccessCard;