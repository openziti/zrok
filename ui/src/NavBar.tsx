import {AppBar, Box, Button, Grid2, Toolbar, Tooltip, Typography} from "@mui/material";
import {useNavigate} from "react-router";
import LogoutIcon from "@mui/icons-material/Logout";
import VisualizerIcon from "@mui/icons-material/Hub";
import TabularIcon from "@mui/icons-material/TableRows";
import LimitIcon from "@mui/icons-material/CrisisAlert";
import HelpIcon from "@mui/icons-material/LiveHelp";
import zrokLogo from "./assets/zrok-1.0.0-rocket-green.svg";
import useApiConsoleStore from "./model/store.ts";
import BandwidthLimitedModal from "./BandwidthLimitedModal.tsx";
import {useEffect, useState} from "react";
import GettingStartedModal from "./GettingStartedModal.tsx";
import {extensionRegistry} from "./extensions/registry.ts";
import {Slot} from "./extensions/SlotRenderer.tsx";
import {SLOTS, ExtensionNavItem} from "./extensions/types.ts";

interface NavBarProps {
    logout: () => void;
    visualizer: boolean;
    toggleMode: (boolean) => void;
}

const NavBar = ({ logout, visualizer, toggleMode }: NavBarProps) => {
    const navigate = useNavigate();
    const user = useApiConsoleStore((state) => state.user);
    const nodes = useApiConsoleStore((state) => state.nodes);
    const limited = useApiConsoleStore((state) => state.limited);
    const extensions = useApiConsoleStore((state) => state.extensions);
    const [limitedModalOpen, setLimitedModalOpen] = useState<boolean>(false);
    const openLimitedModal = () => {
        setLimitedModalOpen(true);
    }
    const closeLimitedModal = () => {
        setLimitedModalOpen(false);
    }
    const [gettingStartedOpen, setGettingStartedOpen] = useState<boolean>(false);
    const openGettingStarted = () => {
        setGettingStartedOpen(true);
    }
    const closeGettingStarted = () => {
        setGettingStartedOpen(false);
    }

    useEffect(() => {
        if(!limited) {
            closeLimitedModal();
        }
    }, [limited])

    // Get extension nav items
    const leftNavItems = extensionRegistry.getNavItems('left');
    const rightNavItems = extensionRegistry.getNavItems('right');

    // Filter visible items based on visibility function
    const filterVisibleItems = (items: Array<ExtensionNavItem & { extensionId: string }>) => {
        return items.filter(item => {
            if (item.visible) {
                const extState = extensions[item.extensionId] || {};
                return item.visible(user, extState);
            }
            return true;
        });
    };

    const visibleLeftItems = filterVisibleItems(leftNavItems);
    const visibleRightItems = filterVisibleItems(rightNavItems);

    // Render a nav item
    const renderNavItem = (item: ExtensionNavItem & { extensionId: string }) => {
        const handleClick = () => {
            if (item.onClick) {
                item.onClick();
            } else if (item.path) {
                navigate(item.path);
            }
        };

        const IconComponent = item.icon;

        return (
            <Grid2 key={`${item.extensionId}-${item.id}`} display="flex" justifyContent="right">
                <Tooltip title={item.tooltip || item.label}>
                    <Button color="inherit" onClick={handleClick}>
                        {IconComponent ? <IconComponent fontSize="medium" /> : item.label}
                    </Button>
                </Tooltip>
            </Grid2>
        );
    };

    const limitedIndicator = (
        <Grid2 display="flex" justifyContent="right">
            <Tooltip title="Bandwidth Limit Reached!">
                <Button color="error" onClick={openLimitedModal}><LimitIcon /></Button>
            </Tooltip>
        </Grid2>
    );

    const gettingStartedButton = (
        <Grid2 display="flex" justifyContent="right">
            <Tooltip title="Getting Started Wizard">
                <Button style={{ backgroundColor: "#9bf316", color: "black" }} onClick={openGettingStarted}>CLICK HERE TO GET STARTED!</Button>
            </Tooltip>
        </Grid2>
    );

    const helpButton = (
        <Grid2 display="flex" justifyContent="right">
            <Tooltip title="Getting Started Wizard">
                <Button style={{ color: "#9bf316" }} onClick={openGettingStarted}><HelpIcon /></Button>
            </Tooltip>
        </Grid2>
    );

    const handleClick = () => {
        toggleMode(!visualizer);
    }

    return (
        <>
            <Box sx={{ flexGrow: 1 }}>
                <AppBar position="static">
                    <Toolbar>
                        <Typography variant="h6" sx={{ flexGrow: 1 }}>
                            <Grid2 container sx={{ flexGrow: 1 }}>
                                <Grid2 display="flex" justifyContent="left">
                                    <img src={zrokLogo} height="30" />
                                </Grid2>
                                <Grid2 display="flex" justifyContent="left" size="grow" sx={{ ml: 3 }} color="#9bf316">
                                    <strong>z r o k</strong>
                                </Grid2>
                                {/* Extension slot: left side of navbar */}
                                <Slot name={SLOTS.NAVBAR_LEFT} user={user} />
                                {/* Extension nav items: left position */}
                                {visibleLeftItems.map(renderNavItem)}
                            </Grid2>
                        </Typography>
                        <Grid2 container sx={{ flexGrow: 1 }}>
                            {/* Extension slot: center of navbar */}
                            <Slot name={SLOTS.NAVBAR_CENTER} user={user} />

                            <Grid2 display="flex" justifyContent="right" size="grow">
                                <Tooltip title="Toggle Interface Mode (Ctrl-`)">
                                    <Button color="inherit" onClick={handleClick}>{ visualizer ? <VisualizerIcon /> : <TabularIcon /> }</Button>
                                </Tooltip>
                            </Grid2>
                            { limited ? limitedIndicator : null }

                            {/* Extension nav items: right position */}
                            {visibleRightItems.map(renderNavItem)}

                            {/* Extension slot: right side of navbar */}
                            <Slot name={SLOTS.NAVBAR_RIGHT} user={user} />

                            { !nodes || nodes.length > 1 ? helpButton : gettingStartedButton }
                            <Grid2 display="flex" justifyContent="right">
                                <Tooltip title="Logout">
                                    <Button color="inherit" onClick={logout}><LogoutIcon /></Button>
                                </Tooltip>
                            </Grid2>
                        </Grid2>
                    </Toolbar>
                </AppBar>
            </Box>
            <BandwidthLimitedModal close={closeLimitedModal} isOpen={limitedModalOpen} />
            <GettingStartedModal close={closeGettingStarted} isOpen={gettingStartedOpen} />
        </>
    );
}

export default NavBar;
