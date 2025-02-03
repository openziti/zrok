import {useState} from "react";
import HideIcon from "@mui/icons-material/Visibility";
import ShowIcon from "@mui/icons-material/VisibilityOff";
import {Grid2} from "@mui/material";

interface SecretToggleProps {
    secret: string;
}

const SecretToggle = ({ secret }: SecretToggleProps) => {
    const [showSecret, setShowSecret] = useState(false);

    const toggle = () => setShowSecret(!showSecret);

    const secretString = (s): string => {
        let out = "";
        for(let i = 0; i < s.length; i++) {
            out += " \u2022"; // bullet
        }
        return out;
    }

    if(showSecret) {
        return (
            <Grid2 container sx={{ flexGrow: 1 }} alignItems="center">
                <Grid2 display="flex" justifyContent="left">
                    <span>{secret}</span>
                </Grid2>
                <Grid2 display="flex" justifyContent="right" sx={{ flexGrow: 1 }}>
                    <HideIcon onClick={toggle} sx={{ ml: 1 }}/>
                </Grid2>
            </Grid2>
        );
    } else {
        return (
            <Grid2 container sx={{ flexGrow: 1 }} alignItems="center">
                <Grid2 display="flex" justifyContent="left">
                    <span>{secretString(secret)}</span>
                </Grid2>
                <Grid2 display="flex" justifyContent="right" sx={{ flexGrow: 1 }}>
                    <ShowIcon onClick={toggle} sx={{ ml: 1 }} />
                </Grid2>
            </Grid2>
        );
    }
}

export default SecretToggle;