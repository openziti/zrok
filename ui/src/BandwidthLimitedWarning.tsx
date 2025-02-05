import {Grid2} from "@mui/material";
import React from "react";

const BandwidthLimitedWarning = () => {
    return (
        <Grid2 container sx={{ flexGrow: 1, mt: 0, mb: 2 }} alignItems="center">
            <h5 style={{ margin: 0, color: "red" }}>Your account is currently over the assigned bandwidth limit!</h5>
        </Grid2>
    );
}

export default BandwidthLimitedWarning;