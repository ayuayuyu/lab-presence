export type User = {
  id: number;
  name: string;
  email: string;
  picture: string;
  student_id: string;
  created_at: string;
};

export type Device = {
  id: number;
  user_id: number;
  mac_address: string;
  label: string;
  created_at: string;
};

export type Presence = {
  user_id: number;
  user_name: string;
  user_picture: string;
  mac_address: string;
  device_label: string;
  detected_at: string;
};

export type LastSeen = {
  user_id: number;
  user_name: string;
  user_picture: string;
  last_seen_at: string;
};
