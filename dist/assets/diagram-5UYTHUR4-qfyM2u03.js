import{p as y}from"./chunk-353BL4L5-C25V1_gO.js";import{_ as l,s as B,g as S,q as F,p as z,a as E,b as P,D as v,H as W,e as D,y as T,E as _,F as A,l as w}from"./index-BAZQiG6o.js";import{p as N}from"./treemap-75Q7IDZK-DhGly7N9.js";import"./semi-ui-CcdwbGAH.js";import"./react-core-BvF_2poO.js";import"./tools-Dh1tmYLu.js";import"./react-components-CT7d80Vq.js";import"./i18n-B-5Xezjc.js";import"./_baseUniq-DOFRIXNf.js";import"./_basePickBy-BzOcr449.js";import"./clone-C7aB-Z04.js";var x={packet:[]},m=structuredClone(x),L=A.packet,Y=l(()=>{const t=v({...L,..._().packet});return t.showBits&&(t.paddingY+=10),t},"getConfig"),H=l(()=>m.packet,"getPacket"),I=l(t=>{t.length>0&&m.packet.push(t)},"pushWord"),M=l(()=>{T(),m=structuredClone(x)},"clear"),u={pushWord:I,getPacket:H,getConfig:Y,clear:M,setAccTitle:P,getAccTitle:E,setDiagramTitle:z,getDiagramTitle:F,getAccDescription:S,setAccDescription:B},O=1e4,q=l(t=>{y(t,u);let e=-1,o=[],i=1;const{bitsPerRow:s}=u.getConfig();for(let{start:a,end:r,bits:c,label:f}of t.blocks){if(a!==void 0&&r!==void 0&&r<a)throw new Error(`Packet block ${a} - ${r} is invalid. End must be greater than start.`);if(a??(a=e+1),a!==e+1)throw new Error(`Packet block ${a} - ${r??a} is not contiguous. It should start from ${e+1}.`);if(c===0)throw new Error(`Packet block ${a} is invalid. Cannot have a zero bit field.`);for(r??(r=a+(c??1)-1),c??(c=r-a+1),e=r,w.debug(`Packet block ${a} - ${e} with label ${f}`);o.length<=s+1&&u.getPacket().length<O;){const[d,p]=G({start:a,end:r,bits:c,label:f},i,s);if(o.push(d),d.end+1===i*s&&(u.pushWord(o),o=[],i++),!p)break;({start:a,end:r,bits:c,label:f}=p)}}u.pushWord(o)},"populate"),G=l((t,e,o)=>{if(t.start===void 0)throw new Error("start should have been set during first phase");if(t.end===void 0)throw new Error("end should have been set during first phase");if(t.start>t.end)throw new Error(`Block start ${t.start} is greater than block end ${t.end}.`);if(t.end+1<=e*o)return[t,void 0];const i=e*o-1,s=e*o;return[{start:t.start,end:i,label:t.label,bits:i-t.start},{start:s,end:t.end,label:t.label,bits:t.end-s}]},"getNextFittingBlock"),K={parse:l(async t=>{const e=await N("packet",t);w.debug(e),q(e)},"parse")},R=l((t,e,o,i)=>{const s=i.db,a=s.getConfig(),{rowHeight:r,paddingY:c,bitWidth:f,bitsPerRow:d}=a,p=s.getPacket(),n=s.getDiagramTitle(),k=r+c,g=k*(p.length+1)-(n?0:r),b=f*d+2,h=W(e);h.attr("viewbox",`0 0 ${b} ${g}`),D(h,g,b,a.useMaxWidth);for(const[C,$]of p.entries())U(h,$,C,a);h.append("text").text(n).attr("x",b/2).attr("y",g-k/2).attr("dominant-baseline","middle").attr("text-anchor","middle").attr("class","packetTitle")},"draw"),U=l((t,e,o,{rowHeight:i,paddingX:s,paddingY:a,bitWidth:r,bitsPerRow:c,showBits:f})=>{const d=t.append("g"),p=o*(i+a)+a;for(const n of e){const k=n.start%c*r+1,g=(n.end-n.start+1)*r-s;if(d.append("rect").attr("x",k).attr("y",p).attr("width",g).attr("height",i).attr("class","packetBlock"),d.append("text").attr("x",k+g/2).attr("y",p+i/2).attr("class","packetLabel").attr("dominant-baseline","middle").attr("text-anchor","middle").text(n.label),!f)continue;const b=n.end===n.start,h=p-2;d.append("text").attr("x",k+(b?g/2:0)).attr("y",h).attr("class","packetByte start").attr("dominant-baseline","auto").attr("text-anchor",b?"middle":"start").text(n.start),b||d.append("text").attr("x",k+g).attr("y",h).attr("class","packetByte end").attr("dominant-baseline","auto").attr("text-anchor","end").text(n.end)}},"drawWord"),X={draw:R},j={byteFontSize:"10px",startByteColor:"black",endByteColor:"black",labelColor:"black",labelFontSize:"12px",titleColor:"black",titleFontSize:"14px",blockStrokeColor:"black",blockStrokeWidth:"1",blockFillColor:"#efefef"},J=l(({packet:t}={})=>{const e=v(j,t);return`
	.packetByte {
		font-size: ${e.byteFontSize};
	}
	.packetByte.start {
		fill: ${e.startByteColor};
	}
	.packetByte.end {
		fill: ${e.endByteColor};
	}
	.packetLabel {
		fill: ${e.labelColor};
		font-size: ${e.labelFontSize};
	}
	.packetTitle {
		fill: ${e.titleColor};
		font-size: ${e.titleFontSize};
	}
	.packetBlock {
		stroke: ${e.blockStrokeColor};
		stroke-width: ${e.blockStrokeWidth};
		fill: ${e.blockFillColor};
	}
	`},"styles"),lt={parser:K,db:u,renderer:X,styles:J};export{lt as diagram};
