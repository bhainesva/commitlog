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
    </div>
  )
}