import ShareIcon from "@mui/icons-material/Share";
import {AppBar, Box, Button, Card, Chip, Grid2, Toolbar, Typography} from "@mui/material";
import DeleteIcon from "@mui/icons-material/Delete";
import {releaseShare} from "./model/handler.js";

const ShareCard = (props) => {
    let frontends = [];
    props.share.frontendEndpoint.map((fe) => {
        frontends.push(<a key={props.share.token} href={fe.toString()} target={"_"}>{fe}</a>);
    })

    const deleteHandler = () => {
        releaseShare({ token: props.share.token });
    }

    return (
        <Card>
            <AppBar position="sticky">
                <Toolbar variant="dense">
                    <ShareIcon />
                    <Grid2 container sx={{ flexGrow: 1 }}>
                        <Grid2 display="flex" justifyContent="center" size="grow">
                            <Typography variant="h6" component="div">{props.share.token}</Typography>
                        </Grid2>
                    </Grid2>
                    <Grid2 container>
                        <Grid2 display="flex" justifyContent="right">
                            {props.share.shareMode === "public" && (
                                <Chip label={props.share.shareMode} size="small" color="success" />
                            )}
                            {props.share.shareMode === "private" && (
                                <Chip label={props.share.shareMode} size="small" color="warning" />
                            )}
                            <Chip label={props.share.backendMode} size="small" color="info" />
                        </Grid2>
                    </Grid2>
                </Toolbar>
            </AppBar>
            <Box sx={{ p: 2, textAlign: "center" }}>
                <Typography variant="h6" component="div">
                    {props.share.backendEndpoint} &rarr; {frontends} <br/>
                </Typography>
            </Box>
            <Grid2 container sx={{ flexGrow: 1 }}>
                <Grid2 display="flex" justifyContent="right" size="grow">
                    <Button variant="contained" onClick={deleteHandler} ><DeleteIcon /></Button>
                </Grid2>
            </Grid2>
        </Card>
    );
}

export default ShareCard;