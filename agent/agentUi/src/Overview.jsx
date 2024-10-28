import "bootstrap/dist/css/bootstrap.min.css";
import LanIcon from "@mui/icons-material/Lan";
import ShareIcon from "@mui/icons-material/Share";
import {Box, Card, Stack} from "@mui/material";
import AccessCard from "./AccessCard.jsx";
import ShareCard from "./ShareCard.jsx";
import React from "react";

const Overview = (props) => {
    let cards = [];
    if(props.overview.length > 0) {
        props.overview.forEach((row) => {
            switch(row.type) {
                case "access":
                    cards.push(<AccessCard key={row.frontendToken} access={row} />);
                    break;

                case "share":
                    cards.push(<ShareCard key={row.token} share={row} />);
                    break;
            }
        });
    } else {
        cards.push(<Card key="empty"><h5>zrok Agent is empty! Add a <a href="#" onClick={props.shareClick}>share <ShareIcon /></a> or <a href={"#"} onClick={props.accessClick}>access <LanIcon /></a> share to get started.</h5></Card>);
    }
    return (
        <Box sx={{ display: "flex",
            flexDirection: "row",
            flexWrap: "wrap",
            justifyContent: "space-between",
            flexGrow: 1
        }}>
            {cards}
        </Box>
    );
}

export default Overview;
