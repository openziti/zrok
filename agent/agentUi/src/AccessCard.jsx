import LanIcon from "@mui/icons-material/Lan";
import {AppBar, Box, Button, Card, Chip, Grid2, Toolbar, Typography} from "@mui/material";
import DeleteIcon from "@mui/icons-material/Delete";
import {releaseAccess} from "./model/handler.js";

const AccessCard = (props) => {
    const deleteHandler = () => {
        releaseAccess({ frontendToken: props.access.frontendToken });
    }

    return (
        <Card>
            <AppBar position="sticky">
                <Toolbar variant="dense">
                    <LanIcon />
                    <Grid2 container sx={{ flexGrow: 1 }}>
                        <Grid2 display="flex" justifyContent="center" size="grow">
                            <Typography variant="h6" component="div">{props.access.frontendToken}</Typography>
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
                    {props.access.token} &rarr; {props.access.bindAddress}
                </Typography>
            </Box>
            <Grid2 container sx={{ flexGrow: 1 }}>
                <Grid2 display="flex" justifyContent="right" size="grow">
                    <Button variant="contained" onClick={deleteHandler}><DeleteIcon /></Button>
                </Grid2>
            </Grid2>
        </Card>
    );
}

export default AccessCard;