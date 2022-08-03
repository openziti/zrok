import Icon from '@mdi/react';
import { mdiLogout } from '@mdi/js';

const logoutIcon = mdiLogout;

const Logout = (props) => {
    const onClick = () => {
        props.logout()
    }

    return (
        <a onClick={onClick}><Icon path={logoutIcon} size={1}/> {props.user.email}</a>
    );
}

export default Logout;
