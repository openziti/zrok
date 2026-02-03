import {AgentObject, categorizeAccesses, categorizeShares} from "./model/overview.ts";
import {Box, Card, Chip, Grid2, Typography} from "@mui/material";
import ShareIcon from "@mui/icons-material/Share";
import LanIcon from "@mui/icons-material/Lan";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import WarningIcon from "@mui/icons-material/Warning";
import ErrorIcon from "@mui/icons-material/Error";
import ShareCard from "./ShareCard.tsx";
import AccessCard from "./AccessCard.tsx";
import {AccessDetail, ShareDetail} from "./api";

interface OverviewProps {
    overview: Array<AgentObject>;
    shareClick: () => void;
    accessClick: () => void;
}

const Overview = ({ overview, shareClick, accessClick }: OverviewProps) => {
    // separate shares and accesses for categorization
    const shares = overview.filter(obj => obj.type === "share").map(obj => obj.v as ShareDetail);
    const accesses = overview.filter(obj => obj.type === "access").map(obj => obj.v as AccessDetail);

    const shareCounts = categorizeShares(shares);
    const accessCounts = categorizeAccesses(accesses);

    const totalActive = shareCounts.active + accessCounts.active;
    const totalRetrying = shareCounts.retrying + accessCounts.retrying;
    const totalFailed = shareCounts.failed + accessCounts.failed;

    const getStatusIcon = (status: string) => {
        switch (status) {
            case 'active': return <CheckCircleIcon fontSize="small" />;
            case 'retrying': return <WarningIcon fontSize="small" />;
            case 'failed': return <ErrorIcon fontSize="small" />;
            default: return null;
        }
    };

    const getStatusColor = (status: string) => {
        switch (status) {
            case 'active': return '#4caf50';
            case 'retrying': return '#ff9800';
            case 'failed': return '#f44336';
            default: return '#9e9e9e';
        }
    };

    const cards = [];
    if(overview.length > 0) {
        overview.forEach((row, index) => {
            switch(row.type) {
                case "access":
                    cards.push(<Grid2 key={`access-${index}`} size={{ xs: 12, md: 6 }}><AccessCard accessObject={row} /></Grid2>);
                    break;

                case "share":
                    cards.push(<Grid2 key={`share-${index}`} size={{ xs: 12, md: 6 }}><ShareCard shareObject={row} /></Grid2>);
                    break;
            }
        });
    } else {
        cards.push(<Grid2 key="empty" size={{ xs: 12 }}>
            <Card>
                <Box sx={{ p: 2, textAlign: "center" }}>
                    <Typography variant="h6" component="div">
                        zrok Agent is empty! Add a <a href={"#"} onClick={shareClick}>share <ShareIcon/></a> or <a
                        href={"#"} onClick={accessClick}>access <LanIcon/></a> share to get started.
                    </Typography>
                </Box>
            </Card>
        </Grid2>);
    }

    return (
        <Grid2 container spacing={2}>
            {overview.length > 0 && (
                <Grid2 size={{ xs: 12 }}>
                    <Card>
                        <Box sx={{ p: 2 }}>
                            <Typography variant="h6" gutterBottom>
                                status summary
                            </Typography>
                            <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
                                {totalActive > 0 && (
                                    <Chip
                                        icon={getStatusIcon('active')}
                                        label={`${totalActive} active`}
                                        size="small"
                                        sx={{ backgroundColor: getStatusColor('active'), color: 'white' }}
                                    />
                                )}
                                {totalRetrying > 0 && (
                                    <Chip
                                        icon={getStatusIcon('retrying')}
                                        label={`${totalRetrying} retrying`}
                                        size="small"
                                        sx={{ backgroundColor: getStatusColor('retrying'), color: 'white' }}
                                    />
                                )}
                                {totalFailed > 0 && (
                                    <Chip
                                        icon={getStatusIcon('failed')}
                                        label={`${totalFailed} failed`}
                                        size="small"
                                        sx={{ backgroundColor: getStatusColor('failed'), color: 'white' }}
                                    />
                                )}
                                {(shares.length > 0 || accesses.length > 0) && (
                                    <Typography variant="body2" sx={{ alignSelf: 'center', ml: 2 }}>
                                        {shares.length} shares, {accesses.length} accesses
                                    </Typography>
                                )}
                            </Box>
                        </Box>
                    </Card>
                </Grid2>
            )}
            {cards}
        </Grid2>
    );
}

export default Overview;