import {Box, Button, Modal, TextField} from "@mui/material";
import {useFormik} from "formik";
import {modalStyle} from "./model/theme.js";
import {createAccess} from "./model/handler.js";

const NewAccessModal = (props) => {
    const newAccessForm = useFormik({
        initialValues: {
            token: "",
            bindAddress: "",
        },
        onSubmit: (v) => {
            createAccess(v);
            props.close();
        },
    });

    return (
        <Modal
            open={props.isOpen}
            onClose={props.close}
        >
            <Box sx={{ ...modalStyle }}>
                <h2>Access...</h2>
                <form onSubmit={newAccessForm.handleSubmit}>
                    <TextField
                        fullWidth
                        id="token"
                        name="token"
                        label="Share Token"
                        value={newAccessForm.values.token}
                        onChange={newAccessForm.handleChange}
                        onBlur={newAccessForm.handleBlur}
                        sx={{ mt: 2 }}
                    />
                    <TextField
                        fullWidth
                        id="bindAddress"
                        name="bindAddress"
                        label="Bind Address"
                        value={newAccessForm.values.bindAddress}
                        onChange={newAccessForm.handleChange}
                        onBlur={newAccessForm.handleBlur}
                        sx={{ mt: 2 }}
                    />
                    <Button color="primary" variant="contained" type="submit" sx={{ mt: 2 }}>Create Access</Button>
                </form>
            </Box>
        </Modal>
    );
}

export default NewAccessModal;