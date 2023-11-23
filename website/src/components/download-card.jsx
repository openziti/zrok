// download-card.jsx
import React from 'react';
import { useAssets } from '@site/src/components/assets-context';
import styles from '@site/src/css/download-card.module.css';
import { useColorMode } from '@docusaurus/theme-common';

const getFilenamePattern = (osName) => {
    switch (osName) {
        case 'Windows':
            return 'windows';
        case 'macOS':
            return 'darwin';
        case 'Linux':
            return 'linux';
        default:
            return '';
    }
};

const DownloadCard = ({ osName, osLogo, infoText, guideLink }) => {
    const { colorMode } = useColorMode();
    const assets = useAssets();
    console.log("Assets in DownloadCard:", assets);
    const filenamePattern = getFilenamePattern(osName);
    const filteredLinks = assets.filter(asset => asset.name.includes(filenamePattern));
    console.log("Filtered assets for", osName, "in DownloadCard:", filteredLinks);

    return (
        // <div className={colorMode === 'dark' ? styles.downloadCardDark : styles.downloadCardLight}>
        <div className={styles.downloadCard}>
            <div className={styles.imgContainer}>
                <img src={osLogo} alt={`${osName} logo`} />
            </div>
            <h3>{osName}</h3>
            {filteredLinks.length > 0 && (
                <ul>
                    {filteredLinks.map((link, index) => (
                        <li key={index}>
                            <a href={link.url}>
                                {link.arch}
                            </a>
                        </li>
                    ))}
                </ul>
            )}
            {guideLink && (
            <div className={styles.cardFooter}>
                <p>{infoText}</p>
                <a href={guideLink}>GUIDE</a>
                <p></p>
            </div>
            )}
        </div>
    );
};

export default DownloadCard;
