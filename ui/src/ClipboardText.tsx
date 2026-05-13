import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import CheckMarkIcon from "@mui/icons-material/Check";
import {useEffect, useRef, useState} from "react";
import {Button, Popover, Typography} from "@mui/material";

interface ClipboardTextProps {
    text: string;
}

const ClipboardText = ({ text }: ClipboardTextProps) => {
    const [copied, setCopied] = useState<boolean>(false);
    const resetTimerRef = useRef<number | null>(null);

    useEffect(() => {
        return () => {
            if (resetTimerRef.current !== null) {
                window.clearTimeout(resetTimerRef.current);
            }
        };
    }, []);

    const scheduleReset = () => {
        if (resetTimerRef.current !== null) {
            window.clearTimeout(resetTimerRef.current);
        }
        resetTimerRef.current = window.setTimeout(() => {
            setCopied(false);
            resetTimerRef.current = null;
        }, 2000);
    };

    const copy = async () => {
        await navigator.clipboard.writeText(text);
        setCopied(true);
        scheduleReset();
    }

    return (
        <>
            <Button onClick={copy} sx={{ minWidth: "30px", color: 'common.black' }}>
                {copied ? <CheckMarkIcon /> : <ContentCopyIcon />}
            </Button>
            <Popover anchorOrigin={{ vertical: "top", horizontal: "right" }} open={copied}><Typography sx={{ p: 2 }}>Copied!</Typography></Popover>
        </>
    );
}

export default ClipboardText;
