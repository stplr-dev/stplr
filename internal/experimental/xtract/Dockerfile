# SPDX-License-Identifier: GPL-3.0-or-later
#
# Stapler
# Copyright (C) 2026 The Stapler Authors
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.

FROM ubuntu:22.04 AS builder

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    tar gzip bzip2 xz-utils zstd \
    zip unzip \
    p7zip-full \
    rar unrar \
    brotli \
    lzip lzma lzop \
    lz4 \
    cpio \
    genisoimage \
    rpm \
    ncompress \
    zlib1g-dev \
    build-essential \
    wget \
    autoconf automake libtool \
    libsnappy-dev \
    && rm -rf /var/lib/apt/lists/*

RUN wget https://github.com/kubo/snzip/archive/refs/tags/v1.0.5.tar.gz -O snzip.tar.gz && \
    tar -xzf snzip.tar.gz && \
    cd snzip-1.0.5 && \
    ./autogen.sh && \
    ./configure && \
    make && \
    make install && \
    cd .. && \
    rm -rf snzip-1.0.5 snzip.tar.gz

WORKDIR /work

# Создаём более сложную структуру payload
RUN mkdir -p payload/subdir/nested && \
    echo "hello archive" > payload/file.txt && \
    echo "data in subdir" > payload/subdir/data.txt && \
    echo "nested content" > payload/subdir/nested/deep.txt && \
    echo "binary data" > payload/binary.bin && \
    dd if=/dev/zero of=payload/subdir/zeros bs=1K count=1 2>/dev/null && \
    ln -s ../file.txt payload/subdir/symlink.txt && \
    ln -s subdir payload/dirlink

RUN chown -R 1000:1000 payload

RUN tar -cf test.tar payload && \
    tar -czf test.tar.gz payload && \
    tar -cjf test.tar.bz2 payload && \
    tar -cJf test.tar.xz payload && \
    tar -czf test.tgz payload && \
    tar -cjf test.tbz payload && \
    tar -cjf test.tbz2 payload && \
    tar -cJf test.txz payload && \
    tar --lzip -cf test.tlz payload && \
    tar -cf temp.tar payload && compress -c temp.tar > test.tar.z && rm temp.tar

RUN tar -cf payload.tar payload

RUN gzip -c payload.tar > test.gz && \
    cp test.gz test.gzip && \
    bzip2 -c payload.tar > test.bz2 && \
    xz -c payload.tar > test.xz && \
    lzma -c payload.tar > test.lzma && \
    lzip -c payload.tar > test.lz && \
    lzip -c payload.tar > test.lzip && \
    lz4 payload.tar test.lz4 && \
    compress -c payload.tar > test.z && \
    compress -c payload.tar > test.Z && \
    zstd payload.tar -o test.zst && \
    cp test.zst test.zstd && \
    brotli payload.tar -o test.br && \
    cp test.br test.brotli

RUN \
    find payload -print | cpio -o -H newc > test.cpio && \
    find payload -print | cpio -o -H newc | gzip > test.cpio.gz && \
    find payload -print | cpio -o -H newc | gzip > test.cpgz

# zlib
RUN echo '#include <zlib.h>' > z.c && \
    echo '#include <stdio.h>' >> z.c && \
    echo 'int main(){FILE*f=fopen("payload.tar","rb");FILE*o=fopen("test.zlib","wb");z_stream s={0};deflateInit(&s,Z_DEFAULT_COMPRESSION);unsigned char in[4096],out[4096];int r;do{r=fread(in,1,4096,f);s.avail_in=r;s.next_in=in;do{s.avail_out=4096;s.next_out=out;deflate(&s,r?Z_NO_FLUSH:Z_FINISH);fwrite(out,1,4096-s.avail_out,o);}while(s.avail_out==0);}while(r);deflateEnd(&s);fclose(f);fclose(o);}' >> z.c && \
    gcc z.c -lz -o z && ./z && rm z z.c

RUN set -e && \
    zip -ry test.zip payload && \
    7z a test.7z payload && \
    7z a -v1k testmulti.7z payload && \
    mv testmulti.7z.001 test.7z.001 && \
    rm -f testmulti.7z.* && \
    rar a -ma4 -vn -v1k -ol test.rar payload

RUN ar r test.ar payload/file.txt payload/binary.bin payload/subdir/data.txt payload/subdir/nested/deep.txt payload/subdir/zeros && \
    ar r test.deb payload/file.txt payload/binary.bin payload/subdir/data.txt payload/subdir/nested/deep.txt payload/subdir/zeros

RUN mkdir -p /root/rpmbuild/BUILD && \
    mkdir -p /root/rpmbuild/RPMS && \
    mkdir -p /root/rpmbuild/SOURCES && \
    mkdir -p /root/rpmbuild/SPECS && \
    mkdir -p /root/rpmbuild/SRPMS && \
    mkdir -p test-1.0 && \
    cp -r payload test-1.0/ && \
    tar -czf /root/rpmbuild/SOURCES/test.tar.gz test-1.0 && \
    rm -rf test-1.0 && \
    cat > /root/rpmbuild/SPECS/test.spec << 'SPECEOF'
Name: test
Version: 1.0
Release: 1
Summary: test
License: MIT
Source0: test.tar.gz

%description
test

%prep
%setup -q

%build

%install
mkdir -p %{buildroot}
cp -r payload %{buildroot}

%files
/payload
SPECEOF

RUN rpmbuild -ba /root/rpmbuild/SPECS/test.spec && \
    cp /root/rpmbuild/RPMS/*/*.rpm /work/test.rpm

RUN cd $(mktemp -d) && cp -r /work/payload payload && genisoimage -o /work/test.iso -R -J . && cd /work

RUN cp test.zlib test.zz && \
    cp test.tar.z test.tz && \
    cp test.tlz test.tar.lz

RUN snzip -c payload.tar > test.sz && \
    cp test.sz test.snappy && \
    cp test.sz test.s2

RUN rm -rf payload payload.tar test-1.0 testmulti.* /root/rpmbuild

FROM scratch
COPY --from=builder --chown=1000:1000 /work/test.* /fixtures/