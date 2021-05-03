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

function Info() {
  return (
    <div className="Info">
      <ul>
        <li>
          Source for the favicon: <a href="https://www.favicon.cc/?action=icon&file_id=934763">https://www.favicon.cc/?action=icon&file_id=934763</a>
        </li>
        <li>
          Concept of 'idealized commit logs' taken from this talk by Alan Shreve: <a href="https://www.youtube.com/watch?v=dSqLt8BgbRQ">https://www.youtube.com/watch?v=dSqLt8BgbRQ</a>
        </li>
        <li>
          Copied a memory cache thing from here: <a href="https://medium.com/@melvinodsa/building-a-high-performant-concurrent-cache-in-golang-b6442c20b2ca">https://medium.com/@melvinodsa/building-a-high-performant-concurrent-cache-in-golang-b6442c20b2ca</a>
        </li>
      </ul>
    </div>
  )
}

const Page: React.FC<PageProps> = (props: PageProps) => {
  const { children, packages, activePackage, hidePicker, onSubmit } = props;

  const [showInfo, setShowInfo] = useState(false);

  return (
    <div className="Page">
      <div className="Header">
        <a href="https://benhaines.dev">
          <img style={{height: '62px', width: '62px'}} src={icon} />
        </a>
        {!hidePicker && <PackagePicker simple={true} packages={packages} onSubmit={onSubmit} />}
        <div style={{textAlign: 'right'}}>
          {activePackage} {activePackage.includes('.') && <a href={`https://pkg.go.dev/${activePackage}`}>pkg.go.dev</a>}
        </div>
      </div>
      <div className="Page-content">
        {showInfo ? Info() : children}
      </div>
      <button className="Page-infoButton" onClick={() => setShowInfo(!showInfo)}>?</button>
    </div>
  )
}

export default Page;