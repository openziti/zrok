import {AgentObject} from "./model/overview.ts";
import {Card, Grid2} from "@mui/material";
import ShareIcon from "@mui/icons-material/Share";
import LanIcon from "@mui/icons-material/Lan";
import ShareCard from "./ShareCard.tsx";
import AccessCard from "./AccessCard.tsx";

interface OverviewProps {
    overview: Array<AgentObject>;
}

function Overview({ overview }: OverviewProps) {
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
                <h5>zrok Agent is empty! Add a <a href={"#"}>share <ShareIcon /></a> or <a href={"#"}>access <LanIcon /></a> share to get started.</h5>
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