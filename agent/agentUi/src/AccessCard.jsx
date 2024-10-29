import LanIcon from "@mui/icons-material/Lan";
import {Button, Card, Chip} from "@mui/material";
import DeleteIcon from "@mui/icons-material/Delete";
import {releaseAccess} from "./model/handler.js";

const AccessCard = (props) => {
    const deleteHandler = () => {
        releaseAccess({ frontendToken: props.access.frontendToken });
    }

    return (
        <Card>
            <h2><LanIcon /> {props.access.frontendToken}</h2>
            <p>
                {props.access.token} &rarr; {props.access.bindAddress}
            </p>
            <Button variant="contained" onClick={deleteHandler}><DeleteIcon /></Button>
        </Card>
    );
}

export default AccessCard;