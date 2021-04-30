
import React from 'react';
import './scss/Toast.scss';

export interface ToastProps {
  msg: string
  err: boolean
}

export default function Toast(props: ToastProps) {
  if (!props) return null;

  const {msg, err} = props;

  return (
    <div className={`Toast Toast--${err ? 'error' : 'success'}`}>
      {msg}
    </div>
  )
}
