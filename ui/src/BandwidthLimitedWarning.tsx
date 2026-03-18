import {Box, Grid2} from "@mui/material";
import React from "react";

const BandwidthLimitedWarning = () => {
    return (
        <Grid2 container sx={{ flexGrow: 1, mt: 0, mb: 2 }} alignItems="center">
            <Box component="h5" sx={{ m: 0, color: 'error.main' }}>Your account is currently over the assigned bandwidth limit!</Box>
        </Grid2>
    );
}

export default BandwidthLimitedWarning;