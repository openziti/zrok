import LanIcon from "@mui/icons-material/Lan";
import DeleteIcon from "@mui/icons-material/Delete";
import {Card} from "@mui/material";

const AccessCard = (props) => {
    const releaseClicked = () => {
        props.releaseAccess({frontendToken: props.access.frontendToken}, (err, data) => { console.log("releaseClicked", data); });
    }

    return (
        <Card sx={{ mt: 2, p: 2 }}>
            <h2>{props.access.frontendToken} [<LanIcon/>]</h2>
            <p>
                {props.access.token} &rarr; {props.access.bindAddress}
            </p>
            <p><DeleteIcon onClick={releaseClicked}/></p>
        </Card>
    );
}

export default AccessCard;