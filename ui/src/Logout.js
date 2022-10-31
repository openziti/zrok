import Icon from '@mdi/react';
import { mdiLogout } from '@mdi/js';

const logoutIcon = mdiLogout;

const Logout = (props) => {
    const onClick = () => {
        props.logout()
    }

    return (
        <button onClick={onClick} aria-label={"log out"} title={"Log out"}><Icon path={logoutIcon} size={.7}/></button>
    );
}

export default Logout;
