import React, { useEffect, useState } from "react";
import './scss/PackagePicker.scss';
import PackagePicker from "./PackagePicker";

interface Props {
  packages: string[],
  onSubmit: (pkg: string) => void
  simple?: boolean,
}

export default function LandingPage(props: Props) {
  const { packages, onSubmit } = props;

  return (
    <div className="LandingPage">
      <PackagePicker packages={packages} onSubmit={onSubmit} />

      <div className="LandingPage-message">
        Start with `commitlog/demo` for a tiny package to demonstrate the concept.<br />

        To use a package not visible here:<br /><br />
        <ul>
          <li>
            Provide an absolute path to a locally cloned package. If it doesn't use go modules it must be inside your $GOPATH.
          </li>
          <li>
            <br />OR<br /><br />
          </li>
          <li>
            `go get` it and then import it into cmd/commitlog-server/imports.go, it will then show up in autocomplete
          </li>
        </ul>
      </div>
    </div>
  )
}