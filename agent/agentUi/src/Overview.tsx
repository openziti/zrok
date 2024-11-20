import {AgentObject} from "./model/overview.ts";
import {Box, Card, Grid2, Typography} from "@mui/material";
import ShareIcon from "@mui/icons-material/Share";
import LanIcon from "@mui/icons-material/Lan";
import ShareCard from "./ShareCard.tsx";
import AccessCard from "./AccessCard.tsx";

interface OverviewProps {
    overview: Array<AgentObject>;
    shareClick: () => void;
    accessClick: () => void;
}

const Overview = ({ overview, shareClick, accessClick }: OverviewProps) => {
    let cards = [];
    if(overview.length > 0) {
        overview.forEach(row => {
            switch(row.type) {
                case "access":
                    cards.push(<Grid2 size={{ xs: 12, md: 6 }}><AccessCard accessObject={row} /></Grid2>);
                    break;

                case "share":
                    cards.push(<Grid2 size={{ xs: 12, md: 6 }}><ShareCard shareObject={row} /></Grid2>);
                    break;
            }
        });
    } else {
        cards.push(<Grid2 size={{ xs: 12 }}>
            <Card key="empty">
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
            {cards}
        </Grid2>
    );
}

export default Overview;