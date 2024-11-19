import {AgentObject} from "./model/overview.ts";
import {AppBar, Box, Button, Card, Chip, Grid2, Toolbar, Typography} from "@mui/material";
import LanIcon from "@mui/icons-material/Lan";
import {AccessDetail} from "./api";
import DeleteIcon from "@mui/icons-material/Delete";

interface AccessCardProps {
    accessObject: AgentObject;
}

function AccessCard({ accessObject }: AccessCardProps) {
    let access = (accessObject.v as AccessDetail);
    return (
        <Card>
            <AppBar position="sticky">
                <Toolbar variant="dense">
                    <LanIcon />
                    <Grid2 container sx={{ flexGrow: 1 }}>
                        <Grid2 display="flex" justifyContent="center" size="grow">
                            <Typography variant="h6" component="div">{access.frontendToken}</Typography>
                        </Grid2>
                    </Grid2>
                    <Grid2 container>
                        <Grid2 display="flex" justifyContent="right">
                            <Chip label="private" size="small" color="warning" />
                        </Grid2>
                    </Grid2>
                </Toolbar>
            </AppBar>
            <Box sx={{ p: 2, textAlign: "center" }}>
                <Typography variant="h6" component="div">
                    {access.token} &rarr; {access.bindAddress}
                </Typography>
            </Box>
            <Grid2 container sx={{ flexGrow: 1 }}>
                <Grid2 display="flex" justifyContent="right" size="grow">
                    <Button variant="contained"><DeleteIcon /></Button>
                </Grid2>
            </Grid2>
        </Card>
    );
}

export default AccessCard;