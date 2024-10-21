import {Box, Button, MenuItem, Modal, TextField} from "@mui/material";
import {useFormik} from "formik";
import {modalStyle} from "./model/theme.js";

const NewShareModal = (props) => {
    const newShareForm = useFormik({
        initialValues: {
            shareMode: "public",
            backendMode: "proxy",
            target: "",
        },
        onSubmit: (v) => {
            props.handler(v);
            props.close();
        },
    });

    return (
        <Modal
            open={props.show}
            onClose={props.close}
        >
            <Box sx={{ ...modalStyle }}>
                <h2>Share...</h2>
                <form onSubmit={newShareForm.handleSubmit}>
                    <TextField
                        fullWidth
                        select
                        id="shareMode"
                        name="shareMode"
                        label="Share Mode"
                        value={newShareForm.values.shareMode}
                        onChange={newShareForm.handleChange}
                        onBlur={newShareForm.handleBlur}
                        sx={{ mt: 2 }}
                    >
                        <MenuItem value="public">public</MenuItem>
                        <MenuItem value="private">private</MenuItem>
                    </TextField>
                    <TextField
                        fullWidth select
                        id="backendMode"
                        name="backendMode"
                        label="Backend Mode"
                        value={newShareForm.values.backendMode}
                        onChange={newShareForm.handleChange}
                        onBlur={newShareForm.handleBlur}
                        sx={{ mt: 2 }}
                    >
                        <MenuItem value="proxy">proxy</MenuItem>
                        <MenuItem value="web">web</MenuItem>
                    </TextField>
                    <TextField
                        fullWidth
                        id="target"
                        name="target"
                        label="Target"
                        value={newShareForm.values.target}
                        onChange={newShareForm.handleChange}
                        onBlur={newShareForm.handleBlur}
                        sx={{ mt: 2 }}
                    />
                    <Button color="primary" variant="contained" type="submit" sx={{ mt: 2 }}>Create Share</Button>
                </form>
            </Box>
        </Modal>
    )
}

export default NewShareModal;