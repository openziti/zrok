import "bootstrap/dist/css/bootstrap.min.css";
import LanIcon from "@mui/icons-material/Lan";
import ShareIcon from "@mui/icons-material/Share";
import {Box, Card, Stack} from "@mui/material";
import AccessCard from "./AccessCard.jsx";
import ShareCard from "./ShareCard.jsx";
import React from "react";
import Grid from '@mui/material/Grid2';

const Overview = (props) => {
    let cards = [];
    if(props.overview.length > 0) {
        props.overview.forEach((row) => {
            switch(row.type) {
                case "access":
                    cards.push(<Grid size={{ xs: 12, md: 6 }}><AccessCard key={row.frontendToken} access={row} /></Grid>);
                    break;

                case "share":
                    cards.push(<Grid size={{ xs: 12, md: 6 }}><ShareCard key={row.token} share={row} /></Grid>);
                    break;
            }
        });
    } else {
        cards.push(<Grid size={{ xs: 12 }}>
            <Card key="empty">
                <h5>zrok Agent is empty! Add a <a href="#" onClick={props.shareClick}>share <ShareIcon /></a> or <a href={"#"} onClick={props.accessClick}>access <LanIcon /></a> share to get started.</h5>
            </Card>
        </Grid>);
    }
    return (
        <Grid container spacing={2}>
            {cards}
        </Grid>
    );
}

export default Overview;
