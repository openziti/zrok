import {useState} from "react";
import {useFormik} from "formik";
import {GetAgentApi} from "./model/api.ts";
import {Box, Button, Modal, TextField, Typography} from "@mui/material";
import {modalStyle} from "./model/theme.ts";

interface NewAccessModalProps {
    close: () => void;
    isOpen: boolean;
}

const NewAccessModal = ({ close, isOpen }: NewAccessModalProps) => {
    const [errorMessage, setErrorMessage] = useState(<></>);

    const newAccessForm = useFormik({
        initialValues: {
            token: "",
            bindAddress: "",
        },
        onSubmit: v => {
            setErrorMessage(<></>);
            GetAgentApi().agentAccessPrivate(v)
                .then(r => {
                    close();
                })
                .catch(e => {
                    e.response.json().then(ex => {
                        setErrorMessage(<p>{ex.message}</p>);
                        console.log(ex.message);
                    })
                });
        }
    });

    return (
          <Modal open={isOpen} onClose={close}>
              <Box sx={{...modalStyle}}>
                  <Typography>
                      <h2>Access...</h2>
                  </Typography>
                  {errorMessage}
                  <form onSubmit={newAccessForm.handleSubmit}>
                      <TextField
                          fullWidth
                          id="token"
                          name="token"
                          label="Share Token"
                          value={newAccessForm.values.token}
                          onChange={newAccessForm.handleChange}
                          onBlur={newAccessForm.handleBlur}
                          sx={{mt: 2}}
                      />
                      <TextField
                          fullWidth
                          id="bindAddress"
                          name="bindAddress"
                          label="Bind Address"
                          value={newAccessForm.values.bindAddress}
                          onChange={newAccessForm.handleChange}
                          onBlur={newAccessForm.handleBlur}
                          sx={{mt: 2}}
                      />
                      <Button color="primary" variant="contained" type="submit" sx={{mt: 2}}>Create Access</Button>
                  </form>
              </Box>
          </Modal>
    );
}

export default NewAccessModal;