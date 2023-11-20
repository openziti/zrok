import React, { createContext, useState, useContext, useEffect } from 'react';

const AssetsContext = createContext([]);

export const useAssets = () => useContext(AssetsContext);

export const AssetsProvider = ({ children }) => {
  const [assets, setAssets] = useState([]);

  useEffect(() => {
    const fetchReleaseAssets = async () => {
      try {
        const response = await fetch('https://api.github.com/repos/openziti/zrok/releases/latest');
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        const filteredAssets = data.assets.map(asset => ({
          name: asset.name,
          url: asset.browser_download_url,
          arch: asset.name.replace('\.tar\.gz','').toUpperCase().split('_')[3]
        }));
        console.log("Fetched assets:", filteredAssets); // Log fetched assets
        setAssets(filteredAssets);
      } catch (error) {
        console.error('Error fetching the release assets:', error);
        // Handle the error state appropriately
      }
    };

    fetchReleaseAssets();
  }, []); // Empty dependency array ensures this runs once after component mounts

  return (
    <AssetsContext.Provider value={assets}>
      {children}
    </AssetsContext.Provider>
  );
};
