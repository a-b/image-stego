Timestamped commit: 9b54dc5b4b912b0c3f5944c1bd7ac008b16beb6e

## Problem statement

There is a method of timestamping digital content (e.g. images) by hashing it and writing the resulting character sequence into a blockchain (e.g. transferring money to a bitcoin address derived from that hash). The moment the  information is written to the blockchain it cannot be altered without changing every subsequent block - which is
technically possible but in reality unfeasible. Therefore, it is a tamper-proof way of claiming intellectual property or proofing existence of any digital content.

Especially if the digital content is an image there is the disadvantage of just needing to change one pixel (actually just one bit) and the resulting hash will be completely different although the original image will be perceptually completely identical.

You won't be able to easily proof that a particular image was created at another date than claimed or that it belonged to someone without having the original image.

I'm imagining an image scraper that looks through social media images and finds out that an image from a e.g. battlefield was actually taken from another battle in another country two years earlier than claimed in the post. The image scraper wouldn't have the original image and the adversary could just change one pixel to deceive the scraper, so it cannot verify the existence.

In reality this use case is exacerbated due to compression, cropping etc. of images.

## One step forward

While the following approach won't solve all the aforementioned problems it may be a step in the right direction and may cause thought for others.

What does the approach do in one sentence:

> It uses steganography to embed merkle tree leaves into chunks of the original image, so that each individual chunk can be verified on its own.

Steganography in this context means using the least significant bits of the image to encode information. There are other techniques but this was the most straight forward for me to implement.

## The Approach






























