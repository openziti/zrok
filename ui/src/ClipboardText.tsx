import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import {useEffect, useState} from "react";
import {Button, Popover, Typography} from "@mui/material";

interface ClipboardTextProps {
    text: string;
}

const ClipboardText = ({ text }: ClipboardTextProps) => {
    const [copied, setCopied] = useState<boolean>(false);
    const [color, setColor] = useState<string>("black");

    useEffect(() => {
        if(copied) {
            setColor("red");
        } else {
            setColor("black");
        }
    }, [copied]);

    const copy = async () => {
        await navigator.clipboard.writeText(text);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    }

    return (
        <>
            <Button onClick={copy} sx={{ minWidth: "30px" }} style={{ color: color }} ><ContentCopyIcon /></Button>
            <Popover anchorOrigin={{ vertical: "top", horizontal: "right" }} open={copied}><Typography sx={{ p: 2 }}>Copied!</Typography></Popover>
        </>
    );
}

export default ClipboardText;
