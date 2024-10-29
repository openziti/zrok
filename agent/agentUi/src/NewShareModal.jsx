import {Box, Button, Checkbox, FormControlLabel, MenuItem, Modal, TextField} from "@mui/material";
import {useFormik} from "formik";
import {modalStyle} from "./model/theme.js";
import {createShare} from "./model/handler.js";

const NewShareModal = (props) => {
    const form = useFormik({
        initialValues: {
            shareMode: "public",
            backendMode: "proxy",
            target: "",
            insecure: false,
        },
        onSubmit: (v) => {
            createShare(v);
            props.close();
        },
    });

    const httpBackendModes = ["proxy", "web", "caddy", "drive"];

    return (
        <Modal
            open={props.isOpen}
            onClose={props.close}
        >
            <Box sx={{ ...modalStyle }}>
                <h2>Share...</h2>
                <form onSubmit={form.handleSubmit}>
                    <TextField
                        fullWidth
                        select
                        id="shareMode"
                        name="shareMode"
                        label="Share Mode"
                        value={form.values.shareMode}
                        onChange={form.handleChange}
                        onBlur={form.handleBlur}
                        sx={{ mt: 2 }}
                    >
                        <MenuItem value="public">public</MenuItem>
                        <MenuItem value="private">private</MenuItem>
                    </TextField>
                    {form.values.shareMode === "public" && (
                        <TextField
                            fullWidth select
                            id="backendMode"
                            name="backendMode"
                            label="Backend Mode"
                            value={form.values.backendMode}
                            onChange={form.handleChange}
                            onBlur={form.handleBlur}
                            sx={{ mt: 2 }}
                        >
                            <MenuItem value="proxy">proxy</MenuItem>
                            <MenuItem value="web">web</MenuItem>
                            <MenuItem value="caddy">caddy</MenuItem>
                            <MenuItem value="drive">drive</MenuItem>
                        </TextField>
                    )}
                    {form.values.shareMode === "private" && (
                        <TextField
                            fullWidth select
                            id="backendMode"
                            name="backendMode"
                            label="Backend Mode"
                            value={form.values.backendMode}
                            onChange={form.handleChange}
                            onBlur={form.handleBlur}
                            sx={{ mt: 2 }}
                        >
                            <MenuItem value="proxy">proxy</MenuItem>
                            <MenuItem value="web">web</MenuItem>
                            <MenuItem value="tcpTunnel">tcpTunnel</MenuItem>
                            <MenuItem value="udpTunnel">udpTunnel</MenuItem>
                            <MenuItem value="caddy">caddy</MenuItem>
                            <MenuItem value="drive">drive</MenuItem>
                            <MenuItem value="socks">socks</MenuItem>
                            <MenuItem value="vpn">vpn</MenuItem>
                        </TextField>
                    )}
                    <TextField
                        fullWidth
                        id="target"
                        name="target"
                        label="Target"
                        value={form.values.target}
                        onChange={form.handleChange}
                        onBlur={form.handleBlur}
                        sx={{ mt: 2 }}
                    />
                    {httpBackendModes.includes(form.values.backendMode) && (
                        <Box>
                            <FormControlLabel
                                control={<Checkbox
                                    id="insecure"
                                    name="insecure"
                                    label="Insecure"
                                    checked={form.values.insecure}
                                    onChange={form.handleChange}
                                    onBlur={form.handleBlur}
                                />}
                                label="Insecure"
                                sx={{ mt: 2 }}
                            />
                        </Box>
                    )}
                    <Button color="primary" variant="contained" type="submit" sx={{ mt: 2 }}>Create Share</Button>
                </form>
            </Box>
        </Modal>
    )
}

export default NewShareModal;