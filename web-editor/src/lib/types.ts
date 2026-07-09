export type CodeFile = {
	name: string;
	content: string;
};

export type ImageFile = {
	name: string;
	content: string; //base64
	height: number;
	width: number;
};

export type ProjectData = {
	name: string;
	codeFiles: CodeFile[];
	images: ImageFile[];
};
