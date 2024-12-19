import {useState} from "react";
import ShowIcon from "@mui/icons-material/Visibility";
import HideIcon from "@mui/icons-material/VisibilityOff";
import {Grid2} from "@mui/material";

interface SecretToggleProps {
    secret: string;
}

const SecretToggle = ({ secret }: SecretToggleProps) => {
    const [showSecret, setShowSecret] = useState(false);

    const toggle = () => setShowSecret(!showSecret);

    if(showSecret) {
        return (
            <Grid2 container sx={{ flexGrow: 1 }} alignItems="center">
                <Grid2 display="flex" justifyContent="left">
                    <span>{secret}</span>
                </Grid2>
                <Grid2 display="flex" justifyContent="right">
                    <HideIcon onClick={toggle} sx={{ ml: 1 }}/>
                </Grid2>
            </Grid2>
        );
    } else {
        return (
            <Grid2 container sx={{ flexGrow: 1 }} alignItems="center">
                <Grid2 display="flex" justifyContent="left">
                    <span>XXXXXXXX</span>
                </Grid2>
                <Grid2 display="flex" justifyContent="right">
                    <ShowIcon onClick={toggle} sx={{ ml: 1 }} />
                </Grid2>
            </Grid2>
        );
    }
}

export default SecretToggle;