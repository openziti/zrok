const Logout = (props) => {
    const onClick = () => {
        props.logout()
    }

    return (
        <button onClick={onClick}>Log Out {props.user.email}</button>
    );
}

export default Logout;
