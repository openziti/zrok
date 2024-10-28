import ShareIcon from "@mui/icons-material/Share";
import {Button, Card} from "@mui/material";
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
            <h2><ShareIcon /> {props.share.token}</h2>
            <p>
                ({props.share.shareMode}, {props.share.backendMode}) <br/>
                {props.share.backendEndpoint} &rarr; {frontends} <br/>
            </p>
            <Button variant="outlined" onClick={deleteHandler} ><DeleteIcon /></Button>
        </Card>
    );
}

export default ShareCard;