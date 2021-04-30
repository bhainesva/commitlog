import React, { FunctionComponent, ReactPropTypes, useEffect, useState } from "react";
import './scss/Page.scss';
import './scss/Header.scss';
import icon from './myfavicon.gif';
import PackagePicker from "./PackagePicker";

interface PageProps {
  children: React.ReactNode
  activePackage?: string
  packages?: string[]
  onSubmit?: (pkg: string) => void
  hidePicker?: boolean
}

const Page: React.FC<PageProps> = (props: PageProps) => {
  const { children, packages, activePackage, hidePicker, onSubmit } = props;
  return (
    <div className="Page">
      <div className="Header">
        <a href="https://www.favicon.cc/?action=icon&file_id=934763">
          <img style={{height: '62px', width: '62px'}} src={icon} />
        </a>
        {!hidePicker && <PackagePicker simple={true} packages={packages} onSubmit={onSubmit} />}
        <div style={{textAlign: 'right'}}>
          {activePackage} {activePackage.includes('.') && <a href={`https://pkg.go.dev/${activePackage}`}>pkg.go.dev</a>}
        </div>
      </div>
      <div className="Page-content">
        {children}
      </div>
    </div>
  )
}

export default Page;