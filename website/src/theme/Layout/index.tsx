import React, { type ReactNode } from 'react';
import { NetFoundryLayout } from '@netfoundry/docusaurus-theme/ui';
import { zrokFooter } from '../../components/footer';

const starProps = {
    repoUrl: 'https://github.com/openziti/zrok',
    label: 'Star zrok on GitHub',
};

export default function Layout({ children }: { children: ReactNode }): ReactNode {
    return (
        <NetFoundryLayout footerProps={zrokFooter} starProps={starProps}>
            {children}
        </NetFoundryLayout>
    );
}
